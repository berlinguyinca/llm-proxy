// Package main provides testing utilities for the proxy server
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"llm-proxy/pkg/config"
	"llm-proxy/pkg/discovery"
	"llm-proxy/pkg/hardware"
	"llm-proxy/pkg/memory"
	"llm-proxy/pkg/normalizer"
	rate_limiter "llm-proxy/pkg/rate_limiter"
	registry "llm-proxy/pkg/registry"
	router "llm-proxy/pkg/router"
)

// DiscoveryEndpoint represents the endpoint configuration for agent discovery
type DiscoveryEndpoint struct {
	BaseURL        string `json:"base_url"`
	ServicePath    string `json:"service_path"`
	DestinationURL string `json:"destination_url,omitempty"` // From discovery services
}

// ModelInfo provides model information for discovery endpoints (wrapper for registry.ModelInfo)
type ModelInfo struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	QualifiedName string          `json:"qualified_name"`
	Device        string          `json:"device"`
	MemorySize    uint64          `json:"memory_size_bytes"`
	Status        registry.Status `json:"status"`
}

const (
	portEnv                 = "PORT"
	defaultPort             = "9999"
	configFile              = "config/models.yaml"
	lmStudioDiscoveryURLEnv = "LM_STUDIO_DISCOVERY_URL"
	defaultDiscoveryURL     = "http://localhost:1234/api/v1/models"
	discoveryEnabledEnv     = "DISCOVERY_ENABLED"
	defaultDiscoveryEnabled = "true"
	apiKeyManagementPrefix  = "LLM_API_KEY_"
)

// Metrics initialization for observability
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_total",
			Help: "Total number of proxied requests by status code and model",
		},
		[]string{"status", "model"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_request_duration_seconds",
			Help:    "Request latency distribution by model",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model"},
	)

	requestCounters = map[string]int64{}

	metricsHandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	}
)

// ModelStatsResponse is the response format for /health endpoint with memory pressure metrics
type ModelStatsResponse struct {
	Status               string               `json:"status"`
	TotalLoaded          int                  `json:"total_loaded"`
	Models               []registry.ModelInfo `json:"models"`
	GPUs                 []hardware.GpuDevice `json:"gpus,omitempty"`
	MemoryUtilization    float64              `json:"memory_utilization,omitempty"`     // Fraction of memory pool used (0.0-1.0)
	SwapPressure         bool                 `json:"swap_pressure,omitempty"`          // True if swap is actively being used
	GpuMemoryUtilization float64              `json:"gpu_memory_utilization,omitempty"` // GPU VRAM utilization if GPU available
	EvictionsPending     int                  `json:"evictions_pending,omitempty"`      // Count of models queued for eviction
}

// MemoryThresholdConfig holds memory threshold configuration
type MemoryThresholdConfig struct {
	ThresholdGB float64 `yaml:"threshold_gb"` // Minimum free memory required in GB
}

func main() {
	portStr := os.Getenv(portEnv)
	if portStr == "" {
		portStr = defaultPort
	}

	addr := fmt.Sprintf(":%s", portStr)
	fmt.Printf("Starting LLM Proxy on http://localhost%s\n", addr)

	// Initialize memory pool manager with 16GB threshold
	memoryThreshold, _ := strconv.ParseFloat(os.Getenv("MEMORY_THRESHOLD_GB"), 64)
	if memoryThreshold == 0 {
		memoryThreshold = 16.0 // Default to 16GB
	}
	fmt.Printf("Memory threshold set to %.1f GB\n", memoryThreshold)

	// Parse rate limit configuration (make completely optional)
	rateLimitMaxTokensStr := os.Getenv("RATE_LIMIT_MAX_TOKENS")
	disableRateLimiting := os.Getenv("DISABLE_RATE_LIMITING") == "true"

	var maxTokens, refillRate float64 = 100.0, 10.0 // Default values

	if !disableRateLimiting && rateLimitMaxTokensStr != "" {
		maxTokens, _ = strconv.ParseFloat(rateLimitMaxTokensStr, 64)
		refillRateStr := os.Getenv("RATE_LIMIT_REFILL_RATE")
		if refillRateStr != "" {
			refillRate, _ = strconv.ParseFloat(refillRateStr, 64)
		}
		fmt.Printf("Rate limiting enabled: %d tokens window, %.1f/sec refill\n", int(maxTokens), refillRate)
	} else if disableRateLimiting {
		fmt.Println("Rate limiting disabled (DISABLE_RATE_LIMITING=true)")
	} else {
		fmt.Printf("Rate limiting enabled with defaults: %d tokens window, %.1f/sec refill\n", int(maxTokens), refillRate)
	}

	memoryPoolManager := memory.NewMemoryPoolManager(uint64(memoryThreshold*1024*1024*1024), memoryThreshold)

	// Detect GPUs for hardware-aware memory management
	gpus, err := hardware.DetectGPUs()
	if err != nil {
		fmt.Printf("No NVIDIA GPUs detected. Running on CPU only.\n")
	} else {
		for _, gpu := range gpus {
			fmt.Printf("  • GPU: %s (%d MiB free)\n", gpu.Name, int64(gpu.MemoryFree)/1024/1024)
		}
	}

	// Discover models from LM Studio if enabled
	discoveryEnabled := os.Getenv(discoveryEnabledEnv) == "true"
	discoveryURL := os.Getenv(lmStudioDiscoveryURLEnv)
	if discoveryURL == "" {
		discoveryURL = defaultDiscoveryURL
	}
	fmt.Printf("Discovery enabled: %v, URL: %s\n", discoveryEnabled, discoveryURL)

	var discoveredModels []discovery.ModelInfo
	if discoveryEnabled {
		fmt.Println("\nDiscovering models from LM Studio...")
		discovered, err := discoverLMStudioModels(discoveryURL)
		if err != nil {
			log.Printf("Warning: Failed to discover models from LM Studio: %v\n", err)
		} else if len(discovered) > 0 {
			fmt.Printf("Discovered %d models from LM Studio\n", len(discovered))
			for _, model := range discovered {
				fmt.Printf("  • %s (%.1f GB)\n", model.Name, float64(model.SizeInBytes)/1024/1024/1024)
			}
		} else {
			fmt.Println("No models discovered from LM Studio")
		}
	}

	// Initialize registry and register models (from config + discovery)
	manager := NewModelManager(memoryPoolManager, gpus, discoveryEnabled, discoveryURL, disableRateLimiting, maxTokens, refillRate)
	if err := manager.LoadFromConfig(); err != nil {
		log.Printf("Error loading configs: %v", err)
	}

	// Auto-load models based on resource hints and available memory
	manager.InitAutoLoad(memoryPoolManager)

	fmt.Println("\n=== Summary ===")
	fmt.Printf("Registered models: %d\n", len(manager.GetModels()))
	fmt.Printf("Discovered models (if discovery enabled): %d\n", len(discoveredModels))

	http.HandleFunc("/health", healthHandler(manager))
	http.Handle("/models/stats", modelStatsHandler(manager))
	http.Handle("/gpu/stats", gpuStatsHandler())
	http.Handle("/models/discover", discoverModelsHandler(manager))

	// Add router for proxying requests
	r := router.NewRouter()
	if err := r.Register("/qwen/", "http://localhost:1234/v1/chat/completions"); err != nil {
		log.Printf("Error registering route: %v", err)
	}
	if err := r.Register("/mistral/", "http://localhost:1234/v1/chat/completions"); err != nil {
		log.Printf("Error registering route: %v", err)
	}
	if err := r.Register("/llama/", "http://localhost:1234/v1/chat/completions"); err != nil {
		log.Printf("Error registering route: %v", err)
	}
	if err := r.Register("/phi/", "http://localhost:1234/v1/chat/completions"); err != nil {
		log.Printf("Error registering route: %v", err)
	}

	http.Handle("/model-", r) // Catch-all for model-specific routes

	// Register Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Add load model endpoint for auto-loading functionality
	http.HandleFunc("/models/load", loadModelHandler(manager))

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\nShutting down proxy...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Unload all models on shutdown
	fmt.Println("\nUnloading all models...")
	manager.UnloadAll()
}

func discoverLMStudioModels(url string) ([]discovery.ModelInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models from LM Studio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LM Studio returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	models, err := discovery.ParseLMStudioModels(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse models: %w", err)
	}

	return models, nil
}

type ModelManager struct {
	registry         *registry.ModelRegistry
	router           *router.Router
	memoryPool       *memory.MemoryPoolManager
	gpus             []hardware.GpuDevice
	discoveryEnabled bool
	discoveryURL     string
	rateLimiter      *rate_limiter.TokenBucketStore // nil if rate limiting is disabled
}

func NewModelManager(pmm *memory.MemoryPoolManager, gpus []hardware.GpuDevice, discoveryEnabled bool, discoveryURL string, disableRateLimiting bool, maxTokens float64, refillRate float64) *ModelManager {
	manager := &ModelManager{
		registry:         registry.NewModelRegistry(),
		router:           router.NewRouter(),
		memoryPool:       pmm,
		gpus:             gpus,
		discoveryEnabled: discoveryEnabled,
		discoveryURL:     discoveryURL,
	}

	// Only initialize rate limiter if not disabled
	if !disableRateLimiting {
		manager.rateLimiter = rate_limiter.NewTokenBucketStore(maxTokens, refillRate)
		fmt.Printf("Rate limiter initialized: %d tokens, %.1f/sec refill\n", int(maxTokens), refillRate)
	} else {
		fmt.Println("Rate limiter disabled")
	}

	return manager
}

func (m *ModelManager) LoadFromConfig() error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("Config file not found: %s\n", configFile)
		return nil
	}

	configData, err := config.LoadModelsFromYAML(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	printModelSummary(configData)

	for _, modelConfig := range configData.Models {
		model := registry.ModelInfo{
			ID:            modelConfig.ID,
			Name:          modelConfig.Name,
			QualifiedName: modelConfig.QualifiedName,
			Device:        "cpu", // Simplified - would use placement logic
			MemorySize:    uint64(modelConfig.SizeGB * 1024 * 1024 * 1024),
			Status:        "unloaded",
		}

		m.registry.RegisterFromConfig(model.ID, model.Name, model.QualifiedName, model.Device, model.MemorySize)

		fmt.Printf("Registered: %s (CPU) - unloaded\n", model.Name)

		// Add to memory pool tracking
		m.memoryPool.AddModel(model.ID, model.MemorySize, model.Device)
	}

	return nil
}

func (m *ModelManager) GetModels() []registry.ModelInfo {
	models := make([]registry.ModelInfo, len(m.registry.GetAll()))
	for i, model := range m.registry.GetAll() {
		models[i] = *model
	}

	// Include discovered models if discovery is enabled
	if m.discoveryEnabled {
		for _, disc := range m.memoryPool.GetModels() {
			model := registry.ModelInfo{
				ID:            disc.ModelName,
				Name:          disc.ModelName,
				QualifiedName: fmt.Sprintf("%s.gguf", disc.ModelName),
				Device:        "unknown",
				MemorySize:    0, // Will be updated when loaded from LM Studio
				Status:        "unloaded",
			}
			m.registry.RegisterFromConfig(model.ID, model.Name, model.QualifiedName, model.Device, model.MemorySize)
		}
	}

	return models
}

func (m *ModelManager) LoadAll(discoveredModels []discovery.ModelInfo) error {
	models := m.GetModels()
	totalLoaded := 0

	for _, model := range models {
		if model.Status == "unloaded" {
			fmt.Printf("\nAttempting to load: %s (%.1f GB)\n", model.Name, float64(model.MemorySize)/1024/1024/1024)

			// For now, simulate loading (in production would use subprocess to run LM Studio load command)
			if m.discoveryEnabled {
				// Model was discovered but not loaded yet
				model.Status = "loaded" // Consider it "loaded" for proxy purposes
				fmt.Printf("  ✓ Loaded (simulated) - Status: %s\n", model.Status)

				m.memoryPool.AddModel(model.ID, model.MemorySize, model.Device)
				totalLoaded++
			} else {
				// Simulate loading time
				time.Sleep(100 * time.Millisecond)
				model.Status = "loaded"
				fmt.Printf("  ✓ Loaded (simulated) - Status: %s\n", model.Status)

				m.memoryPool.AddModel(model.ID, model.MemorySize, model.Device)
				totalLoaded++
			}
		}
	}

	fmt.Printf("\nTotal models loaded: %d\n", totalLoaded)
	return nil
}

func (m *ModelManager) UnloadAll() {
	models := m.GetModels()
	for _, model := range models {
		m.registry.Unload(string(model.Status))
	}
}

// InitAutoLoad initializes automatic model loading based on resource hints and available memory
func (m *ModelManager) InitAutoLoad(pmm *memory.MemoryPoolManager) {
	fmt.Println("\n=== Auto-Load Initialization ===")

	totalMemory := pmm.GetAllPools()[0].Total
	totalUsed := pmm.GetAllPools()[0].Used

	availableMemoryMB := float64(totalMemory-totalUsed) / (1024 * 1024)
	fmt.Printf("Total memory pool: %.0f MB\n", totalMemory/(1024*1024))
	fmt.Printf("Currently used: %.0f MB (%.1f%%)\n", totalUsed, float64(totalUsed)*100/float64(totalMemory))
	fmt.Printf("Available for new loads: %.1f MB\n", availableMemoryMB)

	// Check if models with resource hints should be auto-loaded
	models := m.GetModels()
	for i, model := range models {
		// Look for resource hints in config (would need to access original config data in production)
		// For now, assume we want to load models if enough memory available

		if availableMemoryMB >= 1024 { // If at least 1GB available
			fmt.Printf("\nAuto-loading model %d: %s\n", i+1, model.Name)

			// Simulate loading based on resource hints
			memorySize := float64(model.MemorySize) / (1024 * 1024) // Convert to MB
			if memorySize <= availableMemoryMB {
				model.Status = "loaded"
				pmm.AddModel(model.ID, model.MemorySize, model.Device)
				availableMemoryMB -= memorySize
				fmt.Printf("  ✓ Auto-loaded: %s (%.1f GB)\n", model.Name, memorySize)
			} else {
				fmt.Printf("  - Skipped (insufficient memory: %.1f GB > %.1f GB available)\n", memorySize, availableMemoryMB)
			}
		}
	}

	if availableMemoryMB >= 0 {
		fmt.Printf("\nAuto-load complete. Remaining available memory: %.1f MB\n", availableMemoryMB)
	} else {
		fmt.Printf("\nAuto-load complete. Memory pool is at capacity.\n")
	}
}

func printModelSummary(configs *config.ModelConfigs) {
	fmt.Println("\n=== Loaded Models ===")
	for _, model := range configs.Models {
		deviceType := "CPU" // Simplified - would use placement logic
		fmt.Printf("  • %s\n", model.Name)
		fmt.Printf("    ID:       %s\n", model.ID)
		fmt.Printf("    Device:   CPU (%s)\n", deviceType)
		fmt.Printf("    Size:     %.1f GB\n", model.SizeGB)
		fmt.Printf("    URL:      %s\n", model.URL)
	}
	fmt.Println("====================")
}

func (m *ModelManager) LoadFromAPI(modelID string, url string) error {
	fmt.Printf("Loading model %s from URL: %s\n", modelID, url)

	// Construct load command (would actually spawn subprocess to llama.cpp or similar)
	// For now, we'll simulate successful loading
	model := registry.ModelInfo{
		ID:            modelID,
		Name:          modelID,
		QualifiedName: modelID,
		Device:        "cpu", // Will be updated based on GPU availability
		MemorySize:    0,     // Will be actual size once loaded
		Status:        registry.StatusLoading,
	}

	m.registry.RegisterFromConfig(model.ID, model.Name, model.QualifiedName, model.Device, model.MemorySize)
	model.Status = registry.StatusLoaded
	return nil
}

// RegisterFromConfig registers a model from configuration (for on-demand loading)
func (m *ModelManager) RegisterFromConfig(name, path, device string) (*registry.ModelInfo, error) {
	return m.registry.RegisterModelFromConfig(name, path, device)
}

// LoadFromDisk loads a model directly from disk path (for auto-loading)
func (m *ModelManager) LoadFromDisk(name, path, device string) error {
	return m.registry.LoadFromDisk(name, path, device)
}

func healthHandler(manager *ModelManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models := make([]registry.ModelInfo, len(manager.GetModels()))
		copy(models, manager.GetModels())

		gpus, _ := hardware.DetectGPUs()

		// Get memory utilization from pool manager
		memoryUtilization := 0.0
		evictionsPending := 0
		if manager.memoryPool != nil {
			pools := manager.memoryPool.GetAllPools()
			combinedPool := pools[0] // First pool is the combined total
			if combinedPool.Total > 0 {
				memoryUtilization = float64(combinedPool.Used) / float64(combinedPool.Total)

				// If used exceeds free, there's memory pressure (simulating evictions pending)
				if combinedPool.Used > combinedPool.Free {
					evictionsPending = int(float64(combinedPool.Used-combinedPool.Free) / float64(1024*1024*1024)) // Convert GB difference to model count estimate
					if evictionsPending == 0 {
						evictionsPending = 1
					}
				}
			}
		}

		// Calculate GPU memory utilization if GPU available
		gpuMemoryUtilization := 0.0
		if len(gpus) > 0 && gpus[0].MemoryTotal > 0 {
			gpuMemoryUtilization = float64(gpus[0].MemoryTotal-gpus[0].MemoryFree) / float64(gpus[0].MemoryTotal)
		}

		report := ModelStatsResponse{
			Status:               "healthy",
			TotalLoaded:          len(models),
			Models:               models,
			GPUs:                 gpus,
			MemoryUtilization:    memoryUtilization,
			SwapPressure:         false, // Swap is disabled by default in production
			GpuMemoryUtilization: gpuMemoryUtilization,
			EvictionsPending:     evictionsPending,
		}

		reportBytes, err := json.Marshal(report)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write(reportBytes); writeErr != nil {
			log.Printf("Error writing health response: %v", writeErr)
		}
	}
}

func modelStatsHandler(manager *ModelManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models := manager.GetModels()
		report := ModelStatsResponse{
			Status:      "ok",
			TotalLoaded: len(models),
			Models:      models,
		}

		reportBytes, _ := json.Marshal(report)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(reportBytes); err != nil {
			log.Printf("Error writing model stats response: %v", err)
		}
	}
}

func gpuStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gpus, err := hardware.DetectGPUs()
		errHandled := false

		if len(gpus) > 0 || err != nil {
			reportBytes, _ := json.Marshal(map[string]interface{}{
				"status": "ok",
				"gpus":   gpus,
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, writeErr := w.Write(reportBytes); writeErr != nil {
				log.Printf("Error writing GPU stats response: %v", writeErr)
			}
			errHandled = true
		} else {
			reportBytes, _ := json.Marshal(map[string]interface{}{
				"status": "ok",
				"gpus":   []hardware.GpuDevice{},
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, writeErr := w.Write(reportBytes); writeErr != nil {
				log.Printf("Error writing GPU stats response: %v", writeErr)
			}
			errHandled = true
		}

		_ = errHandled
	}
}

// proxyRequest proxies an HTTP request to a backend service with full body/header support and streaming
func (m *ModelManager) proxyRequest(w http.ResponseWriter, req *http.Request, targetURL string, modelID string) {
	fmt.Printf("Proxying request from /%s -> %s\n", req.URL.Path, targetURL)

	// Apply rate limiting only if enabled (rateLimiter is nil when DISABLE_RATE_LIMITING=true)
	if m.rateLimiter != nil && !m.rateLimiter.Acquire(modelID) {
		log.Printf("Rate limit exceeded for model: %s", modelID)
		http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
		return
	}

	// Parse target URL (handle query params for streaming)
	targetReq := &http.Request{}
	var body io.Reader = nil

	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Redact API keys in logs for security
		if len(bodyBytes) > 0 {
			bodyStr := string(bodyBytes)
			bodyStr = strings.ReplaceAll(bodyStr, "sk-", "***-")
			bodyStr = strings.ReplaceAll(bodyStr, "api-", "***-")
			bodyStr = strings.ReplaceAll(bodyStr, "llm_", "***_")
			if len(bodyStr) < 500 {
				fmt.Printf("Request body preview: %s\n", bodyStr)
			}
		}

		body = bytes.NewBuffer(bodyBytes)
		req.Body = io.NopCloser(body)
	} else if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "stream") {
		// For streaming responses, keep body empty for SSE passthrough
	}

	// Copy headers (exclude hop-by-hop headers)
	headerCopy := make(http.Header)
	for key, values := range req.Header {
		if !isHopByHop(key) {
			headerCopy[key] = values
		}
	}

	targetReq.Header = headerCopy
	targetReq.Method = req.Method
	targetReq.URL = &url.URL{
		Scheme:   "http",
		Opaque:   "",
		Path:     strings.TrimPrefix(req.URL.Path, "/model-") + "/" + strings.TrimSuffix(targetURL, "/"),
		RawQuery: req.URL.RawQuery,
	}

	targetReq.Host = targetURL
	req.Body = io.NopCloser(body)

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(targetReq)
	if err != nil {
		log.Printf("Error proxying request to %s: %v", targetURL, err)
		http.Error(w, "Proxy connection error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		if isHopByHop(key) {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	// Handle streaming responses (passthrough)
	if req.URL.RawQuery == "stream=true" || resp.Header.Get("Content-Type") == "text/event-stream" {
		io.Copy(w, resp.Body)
		return
	}

	// For non-streaming requests, normalize to OpenAI format if needed
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("Error reading response body: %v", readErr)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// Normalize response to OpenAI format
	normalizedBytes, err := normalizer.NormalizeResponse(modelID, modelID, time.Now(), map[string]interface{}{
		"id":     modelID,
		"object": "chat.completion",
		"choices": []map[string]interface{}{{
			"index":   0,
			"message": map[string]interface{}{"role": "assistant", "content": string(bodyBytes)},
		}},
	})
	if err != nil {
		log.Printf("Error normalizing response: %v", err)
		http.Error(w, "Normalization error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, writeErr := w.Write(normalizedBytes); writeErr != nil {
		log.Printf("Error writing response: %v", writeErr)
	}
}

func isHopByHop(header string) bool {
	lower := strings.ToLower(header)
	hopHeaders := []string{"connection", "content-length", "transfer-encoding", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "upgrade"}
	for _, h := range hopHeaders {
		if lower == h {
			return true
		}
	}
	return false
}

// loadModelHandler loads a model from disk or remote source
func loadModelHandler(manager *ModelManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req struct {
			Name   string `json:"name"`
			Path   string `json:"path,omitempty"`
			Device string `json:"device,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "Model name is required", http.StatusBadRequest)
			return
		}

		// Register the model in the registry (if not already present)
		model, err := manager.RegisterFromConfig(req.Name, req.Path, req.Device)
		if err != nil {
			log.Printf("Error registering model: %v", err)
			http.Error(w, fmt.Sprintf("Failed to register model: %v", err), http.StatusInternalServerError)
			return
		}

		// Load the model from disk
		if err := manager.LoadFromDisk(model.ID, model.URL, model.Device); err != nil {
			log.Printf("Error loading model %s: %v", req.Name, err)
			http.Error(w, fmt.Sprintf("Failed to load model: %v", err), http.StatusServiceUnavailable)
			return
		}

		log.Printf("Successfully loaded model: %s (%s) from %s", req.Name, model.Device, model.URL)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"model":   model.ID,
			"name":    model.Name,
			"device":  model.Device,
			"url":     model.URL,
			"status":  string(model.Status),
		})
	}
}

// discoverModelsHandler exposes model registry information for agent discovery (Opencode integration)
func discoverModelsHandler(manager *ModelManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		models := manager.GetModels()

		modelsWrapped := make([]ModelInfo, len(models))
		for i, m := range models {
			modelsWrapped[i] = ModelInfo{
				ID:            m.ID,
				Name:          m.Name,
				QualifiedName: m.QualifiedName,
				Device:        m.Device,
				MemorySize:    m.MemorySize,
				Status:        registry.Status(m.Status),
			}
		}

		discoveryResponse := struct {
			ServiceName     string      `json:"service_name"`
			Version         string      `json:"version"`
			Description     string      `json:"description"`
			ModelCount      int         `json:"model_count"`
			Models          []ModelInfo `json:"models"`
			EndpointBaseURL string      `json:"endpoint_base_url"`
			EndpointPath    string      `json:"endpoint_path"`
		}{
			ServiceName:     "llm-proxy",
			Version:         "1.0.0",
			Description:     "LLM Proxy model registry for agent integration and discovery",
			ModelCount:      len(modelsWrapped),
			Models:          modelsWrapped,
			EndpointBaseURL: fmt.Sprintf("http://%s", r.Host),
			EndpointPath:    "/models/stats",
		}

		responseBytes, err := json.MarshalIndent(discoveryResponse, "", "  ")
		if err != nil {
			log.Printf("Error marshaling discovery response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write(responseBytes); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
	}
}
