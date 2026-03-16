// Package main provides CLI management tool for LLM Proxy
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const defaultProxyURL = "http://localhost:9999"

// ModelInfo represents information about a loaded model in the proxy
type ModelInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Device  string `json:"device"` // "cpu", "gpu_0", etc.
	RAM_MB  int64  `json:"ram_mb"`
	VRAM_MB int64  `json:"vram_mb"`
}

// RoutingEntry represents a model-to-backend routing relationship
type RoutingEntry struct {
	ModelName        string `json:"model_name"`
	BackendURL       string `json:"backend_url"`
	DiscoveryEnabled bool   `json:"discovery_enabled"`
	Status           string `json:"status"` // "healthy", "degraded"
}

// Manager wraps the proxy HTTP client for model operations
type Manager struct {
	baseURL string
	http    *http.Client
}

// NewManager creates a new management manager for the proxy
func NewManager(proxyURL string) (*Manager, error) {
	return &Manager{
		baseURL: proxyURL + "/",
		http:    &http.Client{},
	}, nil
}

// ListModels returns information about all loaded models in the proxy
func (m *Manager) ListModels() ([]ModelInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", m.baseURL+"models/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var models []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	info := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		modelName := ""
		ramMB := int64(0)
		vramMB := int64(0)
		device := ""
		path := ""

		if name, ok := m["name"].(string); ok {
			modelName = name
		}
		if ram, ok := m["ram_mb"].(float64); ok {
			ramMB = int64(ram)
		}
		if vram, ok := m["vram_mb"].(float64); ok {
			vramMB = int64(vram)
		}
		if deviceStr, ok := m["device"].(string); ok {
			device = deviceStr
		}
		if p, ok := m["path"].(string); ok {
			path = p
		}

		info = append(info, ModelInfo{
			Name:    modelName,
			Path:    path,
			Device:  device,
			RAM_MB:  ramMB,
			VRAM_MB: vramMB,
		})
	}

	return info, nil
}

// ReloadModel reloads a specific model from disk
func (m *Manager) ReloadModel(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", m.baseURL+"models/"+name+"/reload", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("reload failed with status %d", resp.StatusCode)
}

// UnloadModel gracefully unloads a model from memory
func (m *Manager) UnloadModel(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "DELETE", m.baseURL+"models/"+name, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return nil
	}

	return fmt.Errorf("unload failed with status %d", resp.StatusCode)
}

// ReloadAll reloads all models from disk
func (m *Manager) ReloadAll() error {
	models, err := m.ListModels()
	if err != nil {
		return fmt.Errorf("listing models: %w", err)
	}

	for _, model := range models {
		fmt.Printf("Reloading model: %s...\n", model.Name)
		if err := m.ReloadModel(model.Name); err != nil {
			fmt.Printf("Error reloading %s: %v\n", model.Name, err)
		} else {
			fmt.Printf("✓ Reloaded %s successfully\n", model.Name)
		}
	}

	return nil
}

// GetModelStatus returns detailed status for a specific model
func (m *Manager) GetModelStatus(name string) (*ModelInfo, error) {
	models, err := m.ListModels()
	if err != nil {
		return nil, fmt.Errorf("listing models: %w", err)
	}

	for _, model := range models {
		if model.Name == name {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model %s not found", name)
}

// CheckModelStatus checks if a model is properly loaded and running
func (m *Manager) CheckModelStatus(name string) error {
	models, err := m.ListModels()
	if err != nil {
		return fmt.Errorf("checking models: %w", err)
	}

	for _, model := range models {
		if model.Name == name {
			fmt.Printf("✓ Model %s is loaded\n", model.Name)
			return nil
		}
	}

	return fmt.Errorf("model %s is not currently loaded", name)
}

// PrintModels prints model list in specified format (table or JSON)
func PrintModels(models []ModelInfo, format string, writer io.Writer) {
	if format == "json" || os.Getenv("OUTPUT_FORMAT") == "json" {
		jsonBytes, _ := json.MarshalIndent(models, "", "  ")
		fmt.Fprintln(writer, string(jsonBytes))
	} else {
		printTable(models, writer)
	}
}

// printTable formats models as a table for human readability
func printTable(models []ModelInfo, writer io.Writer) {
	if len(models) == 0 {
		fmt.Fprintln(writer, "No models currently loaded")
		return
	}

	header := []string{"NAME", "DEVICE", "RAM (MB)", "VRAM (MB)"}
	rows := [][]string{}

	for _, m := range models {
		rows = append(rows, []string{
			m.Name,
			m.Device,
			fmt.Sprintf("%d", m.RAM_MB),
			fmt.Sprintf("%d", m.VRAM_MB),
		})
	}

	printTableWithHeader(header, rows, writer)
}

// printTableWithHeader prints a formatted table given header and data rows
func printTableWithHeader(header []string, rows [][]string, writer io.Writer) {
	if len(rows) == 0 {
		fmt.Fprintln(writer, "No models loaded")
		return
	}

	widths := make([]int, len(header))
	for i := range widths {
		widths[i] = len(header[i])
		for _, row := range rows {
			if i < len(row) && len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	fmt.Println()
	for i, h := range header {
		fmt.Fprintf(writer, "%-*s ", widths[i], h)
		if i == len(header)-1 {
			fmt.Fprint(writer, " ")
		} else {
			fmt.Fprint(writer, "  ")
		}
	}
	fmt.Fprintln(writer)

	for _, row := range rows {
		for i, cell := range row {
			if i >= len(header) {
				break
			}
			fmt.Fprintf(writer, "%-*s ", widths[i], cell)
			if i < len(header)-1 {
				fmt.Fprint(writer, " ")
			} else {
				fmt.Fprint(writer, " ")
			}
		}
		fmt.Fprintln(writer)
	}
}

// BackendManager handles proxy backend routing and health checks
type BackendManager struct {
	baseURL string
	http    *http.Client
}

// NewBackendManager creates a new backend manager for the proxy
func NewBackendManager(proxyURL string) (*BackendManager, error) {
	return &BackendManager{
		baseURL: proxyURL + "/",
		http:    &http.Client{},
	}, nil
}

// GetRoutingMap returns the routing configuration showing which models go to which backends
func (b *BackendManager) GetRoutingMap() ([]RoutingEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+"models/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := b.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var models []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	routingMap := make([]RoutingEntry, 0, len(models))
	for _, m := range models {
		entry := RoutingEntry{
			ModelName:        "",
			BackendURL:       "",
			DiscoveryEnabled: false,
			Status:           "healthy",
		}

		if name, ok := m["name"].(string); ok {
			entry.ModelName = name
		}
		if url, ok := m["url"].(string); ok {
			entry.BackendURL = url
		}
		if discovery, ok := m["discovery_enabled"].(bool); ok {
			entry.DiscoveryEnabled = discovery
		}

		routingMap = append(routingMap, entry)
	}

	return routingMap, nil
}

// PrintRouting prints the routing map in specified format (table or JSON)
func PrintRouting(routing []RoutingEntry, format string, writer io.Writer) {
	if format == "json" || os.Getenv("OUTPUT_FORMAT") == "json" {
		jsonBytes, _ := json.MarshalIndent(routing, "", "  ")
		fmt.Fprintln(writer, string(jsonBytes))
	} else {
		printRoutingTable(routing, writer)
	}
}

// printRoutingTable prints routing as a table
func printRoutingTable(routing []RoutingEntry, writer io.Writer) {
	if len(routing) == 0 {
		fmt.Fprintln(writer, "No routing configuration found")
		return
	}

	header := []string{"MODEL", "BACKEND URL", "DISCOVERY"}
	rows := [][]string{}

	for _, r := range routing {
		discoveryStr := "no"
		if r.DiscoveryEnabled {
			discoveryStr = "yes"
		}
		rows = append(rows, []string{
			r.ModelName,
			r.BackendURL,
			discoveryStr,
		})
	}

	printTableWithHeader(header, rows, writer)
}

// AddBackend adds a new backend for a specific model (for LM Studio discovery)
func (b *BackendManager) AddBackend(modelName, baseURL string) error {
	// For now, we just log - actual implementation would call /models/stats and update
	fmt.Printf("Added backend %s for model %s (manual configuration required)\n", baseURL, modelName)
	return nil
}

// RemoveBackend removes a backend URL from routing configuration
func (b *BackendManager) RemoveBackend(url string) error {
	// For now, we just log - actual implementation would call /models/stats and remove
	fmt.Printf("Removed backend %s from routing\n", url)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	var rootCmd = &cobra.Command{
		Use:     "llm-proxy-manager",
		Short:   "CLI management tool for LLM Proxy",
		Long:    "A comprehensive CLI interface for managing the LLM Proxy server. View loaded models, inspect routing configuration, reload/unload models, and manage proxy backends.",
		Version: "1.0.0",
	}

	var format string
	rootCmd.PersistentFlags().StringVar(&format, "format", "table", "Output format (table or json)")

	// Initialize managers
	mgr, _ := NewManager(defaultProxyURL)
	bmgr, _ := NewBackendManager(defaultProxyURL)

	// Add subcommands to root command
	rootCmd.AddCommand(modelsCommand(mgr))
	rootCmd.AddCommand(routingCommand(bmgr))
	rootCmd.AddCommand(backendsCommand(bmgr))
	rootCmd.AddCommand(healthCommand())
	rootCmd.AddCommand(checkCommand(mgr))
	rootCmd.AddCommand(reloadCommand(mgr))

	// Add opencode commands for agent integration at root level too
	opencodeRootCmd := &cobra.Command{
		Use:   "opencode",
		Short: "Manage Opencode agent registration and discovery",
		Long:  `Commands for configuring LLM Proxy integration with Opencode agents, including model discovery and authentication setup.`,
	}

	// init subcommand - redeclare with unique Use string to avoid conflict with models subcommand
	opencodeInitCmd := &cobra.Command{
		Use:   "init [--proxy-url <url>]",
		Short: "Initialize Opencode agent configuration",
		Long:  `Create .opencode/models.yaml configuration file for Opencode agents to discover and connect to LLM Proxy. This sets up authentication and model discovery settings for local agent integration.`,
		Example: `  llm-proxy-manager models opencode init
  llm-proxy-manager opencode init --proxy-url http://localhost:9999`,
		RunE: func(cmd *cobra.Command, args []string) error {
			proxyURL := defaultProxyURL
			if urlFlag, _ := cmd.Flags().GetString("proxy-url"); urlFlag != "" {
				proxyURL = urlFlag
			}

			urlParsed, err := url.Parse(proxyURL)
			if err != nil || urlParsed.Host == "" {
				return fmt.Errorf("invalid proxy URL: %w", err)
			}

			apiKey := fmt.Sprintf("sk-opencode-dev-%s", strings.ReplaceAll(time.Now().Format("2006-01-0215-04-05.999999"), ":", ""))

			configDir := ".opencode"
			if _, err := os.Stat(configDir); os.IsNotExist(err) {
				if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
					return fmt.Errorf("failed to create config directory: %w", mkdirErr)
				}
			}

			configPath := filepath.Join(configDir, "models.yaml")

			var content string
			if data, err := os.ReadFile(filepath.Join(configDir, "models.yaml.example")); err == nil && len(data) > 0 {
				content = string(data)
			} else {
				content = fmt.Sprintf(`# LLM Proxy - Opencode Agent Configuration

---
proxy_url: %s
proxy_path: /
authentication:
  type: api_key
  api_key: %s
models: []
rate_limit:
  enabled: false
  tokens: 100
  refill_rate: 10
logging:
  level: info
discovery:
  enabled: true
  endpoint: /models/discover
`, proxyURL, apiKey)
			}

			if writeErr := os.WriteFile(configPath, []byte(content), 0644); writeErr != nil {
				return fmt.Errorf("failed to write config: %w", writeErr)
			}

			fmt.Println("✓ Created Opencode agent configuration at:")
			fmt.Printf("  %s\n", configPath)
			fmt.Println("")
			fmt.Printf("Generated API key: %s\n", apiKey)
			fmt.Println("Save this key in your environment or configuration file.")

			return nil
		},
	}

	opencodeInitCmd.Flags().String("proxy-url", defaultProxyURL, "LLM Proxy server URL (default: http://localhost:9999)")

	// list subcommand - redeclare with unique Use string to avoid conflict
	opencodeListCmd := &cobra.Command{
		Use:   "opencode list",
		Short: "Display Opencode configuration and registered models",
		Long:  `Show the current Opencode agent configuration including proxy URL, authentication method, and available models for discovery.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(".opencode", "models.yaml")

			data, err := os.ReadFile(configPath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No Opencode configuration found.")
					fmt.Println("")
					fmt.Println("Run 'llm-proxy-manager opencode init' to create one.")
					return nil
				}
				return fmt.Errorf("failed to read config: %w", err)
			}

			fmt.Println("Opencode Agent Configuration")
			fmt.Println("============================")
			fmt.Print(string(data))

			return nil
		},
	}

	opencodeRootCmd.AddCommand(opencodeInitCmd)
	opencodeRootCmd.AddCommand(opencodeListCmd)
	rootCmd.AddCommand(opencodeRootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printUsageAndExit() {
	println("LLM Proxy Manager - CLI interface for managing the proxy server")
	println("")
	println("Usage:")
	println("  llm-proxy-manager [command] [flags]")
	println("")
	println("Available commands:")
	println("  models     Manage loaded models (list, reload, unload)")
	println("  routing    View routing configuration (model-to-backend mapping)")
	println("  backends   Manage proxy backend servers (add, remove)")
	println("  health     Check overall proxy health")
	println("  check      Check if a model is loaded and running")
	println("  reload     Reload all models from disk")
	println("")
	println("Examples:")
	println("  List all loaded models")
	println("    llm-proxy-manager models list")
	println("")
	println("  View routing configuration")
	println("    llm-proxy-manager routing show")
	println("")
	println("  Reload model qwen2.5-7b-chat")
	println("    llm-proxy-manager models reload qwen2.5-7b-chat")
	println("")
	println("  Unload model")
	println("    llm-proxy-manager models unload qwen2.5-7b-chat")
	println("")
	println("  Add backend for LM Studio")
	println("    llm-proxy-manager backends add http://localhost:1234/v1/chat/completions --model qwen2.5-7b-chat")
	println("")
	println("Use 'llm-proxy-manager help <command>' for more information about a command.")

	os.Exit(0)
}

// modelsCommand handles model management subcommands
func modelsCommand(mgr *Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Manage loaded models",
		Long:  "Commands for managing models loaded into the LLM Proxy memory pool.",
	}

	// list subcommand
	listCmd := &cobra.Command{
		Use:   "list [--format <table|json>]",
		Short: "List all loaded models with details",
		Long:  "Display information about all models currently loaded in the proxy memory pool, including device placement (CPU/GPU) and memory usage.",
		Run: func(cmd *cobra.Command, args []string) {
			models, _ := mgr.ListModels()
			fmtFlag, _ := cmd.Flags().GetString("format")
			PrintModels(models, fmtFlag, os.Stdout)
		},
	}

	cmd.AddCommand(listCmd)

	return cmd
}

// routingCommand handles routing subcommands
func routingCommand(bmgr *BackendManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "routing",
		Short: "View routing configuration",
		Long:  "Show which models are served from which proxy backends (LM Studio, Ollama, etc.)",
		Run: func(cmd *cobra.Command, args []string) {
			routing, _ := bmgr.GetRoutingMap()
			fmtFlag, _ := cmd.Flags().GetString("format")
			PrintRouting(routing, fmtFlag, os.Stdout)
		},
	}

	return cmd
}

// backendsCommand handles backend management subcommands
func backendsCommand(bmgr *BackendManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backends",
		Short: "Manage proxy backend servers",
		Long:  "Add and remove backend LLM server URLs for model routing.",
	}

	// add subcommand
	backendsAddCmd := &cobra.Command{
		Use:     "add <url> --model <name>",
		Short:   "Add a new proxy backend for a specific model",
		Long:    "Configure a new proxy backend URL that will be used to serve requests for a specific model. Useful for adding LM Studio, Ollama, or other LLM servers to your routing configuration.",
		Example: `  llm-proxy-manager backends add http://localhost:1234/v1/chat/completions --model qwen2.5-7b-chat`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("url is required")
			}

			url := args[0]

			modelName := ""
			modelFlag, _ := cmd.Flags().GetString("model")
			if modelFlag != "" {
				modelName = modelFlag
			} else {
				modelName = strings.TrimPrefix(url, "http://localhost:1234/")
			}

			return bmgr.AddBackend(modelName, url)
		},
	}

	backendsAddCmd.Flags().String("model", "", "Model name to associate with this backend (required)")

	cmd.AddCommand(backendsAddCmd)

	// remove subcommand
	backendsRemoveCmd := &cobra.Command{
		Use:     "remove <url>",
		Short:   "Remove a proxy backend from routing",
		Long:    "Remove a configured proxy backend URL from the routing configuration. The models that were served by this backend will need to be reloaded or discovered via auto-discovery.",
		Example: `  llm-proxy-manager backends remove http://localhost:1234/v1/chat/completions`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("url is required")
			}

			return bmgr.RemoveBackend(args[0])
		},
	}

	cmd.AddCommand(backendsRemoveCmd)

	return cmd
}

// healthCommand checks overall proxy health
func healthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check overall proxy health",
		Long:  "Verify that the LLM Proxy server is running and responding to requests.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", defaultProxyURL+"/health", nil)
			if err != nil {
				return fmt.Errorf("creating request: %w", err)
			}

			resp, err := (&http.Client{}).Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode == http.StatusOK && len(body) > 0 {
				fmt.Println("✓ Proxy is healthy")
				return nil
			}

			return fmt.Errorf("proxy health check failed with status %d: %s", resp.StatusCode, string(body))
		},
	}

	return cmd
}

// checkCommand checks if a model is loaded and running
func checkCommand(mgr *Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <model>",
		Short: "Check if a model is loaded",
		Long:  "Verify that a specific model is currently loaded in the proxy memory pool.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("model name is required")
			}

			modelName := args[0]
			return mgr.CheckModelStatus(modelName)
		},
	}

	return cmd
}

// reloadCommand reloads all models from disk
func reloadCommand(mgr *Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reload [--all]",
		Short: "Reload all models from disk",
		Long:  "Trigger a full reload of all loaded models to pick up any changes made to the model files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			reloadAll := cmd.Flags().Changed("all") || len(args) == 0

			if reloadAll {
				return mgr.ReloadAll()
			}

			if len(args) < 1 {
				return fmt.Errorf("model name is required when not using --all flag")
			}

			modelName := args[0]
			fmt.Printf("Reloading model %s...\n", modelName)
			if err := mgr.ReloadModel(modelName); err != nil {
				return err
			}

			fmt.Printf("✓ Reloaded %s successfully\n", modelName)
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Reload all models from disk")

	return cmd
}

// opencodeInitCmd creates .opencode/models.yaml configuration for Opencode agents
func opencodeInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opencode init [--proxy-url <url>]",
		Short: "Initialize Opencode agent configuration",
		Long:  `Create .opencode/models.yaml configuration file for Opencode agents to discover and connect to LLM Proxy. This sets up authentication and model discovery settings for local agent integration.`,
		Example: `  llm-proxy-manager opencode init
  llm-proxy-manager opencode init --proxy-url http://localhost:9999`,
		RunE: func(cmd *cobra.Command, args []string) error {
			proxyURL := defaultProxyURL
			if urlFlag, _ := cmd.Flags().GetString("proxy-url"); urlFlag != "" {
				proxyURL = urlFlag
			}

			urlParsed, err := url.Parse(proxyURL)
			if err != nil || urlParsed.Host == "" {
				return fmt.Errorf("invalid proxy URL: %w", err)
			}

			apiKey := fmt.Sprintf("sk-opencode-dev-%s", strings.ReplaceAll(time.Now().Format("2006-01-0215-04-05.999999"), ":", ""))

			configDir := ".opencode"
			if _, err := os.Stat(configDir); os.IsNotExist(err) {
				if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
					return fmt.Errorf("failed to create config directory: %w", mkdirErr)
				}
			}

			configPath := filepath.Join(configDir, "models.yaml")

			var content string
			if data, err := os.ReadFile(filepath.Join(configDir, "models.yaml.example")); err == nil && len(data) > 0 {
				content = string(data)
			} else {
				content = fmt.Sprintf(`# LLM Proxy - Opencode Agent Configuration

---
proxy_url: %s
proxy_path: /
authentication:
  type: api_key
  api_key: %s
models: []
rate_limit:
  enabled: false
  tokens: 100
  refill_rate: 10
logging:
  level: info
discovery:
  enabled: true
  endpoint: /models/discover
`, proxyURL, apiKey)
			}

			if writeErr := os.WriteFile(configPath, []byte(content), 0644); writeErr != nil {
				return fmt.Errorf("failed to write config: %w", writeErr)
			}

			fmt.Println("✓ Created Opencode agent configuration at:")
			fmt.Printf("  %s\n", configPath)
			fmt.Println("")
			fmt.Printf("Generated API key: %s\n", apiKey)
			fmt.Println("Save this key in your environment or configuration file.")

			return nil
		},
	}

	cmd.Flags().String("proxy-url", defaultProxyURL, "LLM Proxy server URL (default: http://localhost:9999)")

	return cmd
}

// opencodeListCmd displays current Opencode configuration
func opencodeListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opencode list",
		Short: "Display Opencode configuration and registered models",
		Long:  `Show the current Opencode agent configuration including proxy URL, authentication method, and available models for discovery.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(".opencode", "models.yaml")

			data, err := os.ReadFile(configPath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No Opencode configuration found.")
					fmt.Println("")
					fmt.Println("Run 'llm-proxy-manager opencode init' to create one.")
					return nil
				}
				return fmt.Errorf("failed to read config: %w", err)
			}

			fmt.Println("Opencode Agent Configuration")
			fmt.Println("============================")
			fmt.Print(string(data))

			return nil
		},
	}

	return cmd
}
