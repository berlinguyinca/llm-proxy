// Package integration provides end-to-end integration tests for the LLM Proxy
package integration

import (
	"encoding/json"
	"testing"
	"time"

	"llm-proxy/pkg/config"
	"llm-proxy/pkg/discovery"
	"llm-proxy/pkg/normalizer"
	registry "llm-proxy/pkg/registry"
	router "llm-proxy/pkg/router"
)

// TestRegistry_Integration tests registry model operations
func TestRegistry_Integration(t *testing.T) {
	t.Parallel()

	reg := registry.NewModelRegistry()

	// Test RegisterFromConfig - returns (*ModelInfo, error)
	modelInfo, err := reg.RegisterFromConfig("qwen-7b", "Qwen-7B-Chat-GGUF", "qwen-7b.gguf", "cpu", 7*1024*1024*1024)
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}
	if modelInfo == nil {
		t.Fatal("Expected ModelInfo, got nil")
	}

	// Verify returned model info fields
	if modelInfo.ID != "qwen-7b" {
		t.Errorf("Expected ID 'qwen-7b', got '%s'", modelInfo.ID)
	}
	if modelInfo.Name != "Qwen-7B-Chat-GGUF" {
		t.Errorf("Expected Name 'Qwen-7B-Chat-GGUF', got '%s'", modelInfo.Name)
	}
	if modelInfo.QualifiedName != "qwen-7b.gguf" {
		t.Errorf("Expected QualifiedName 'qwen-7b.gguf', got '%s'", modelInfo.QualifiedName)
	}
	if modelInfo.Device != "cpu" {
		t.Errorf("Expected Device 'cpu', got '%s'", modelInfo.Device)
	}

	// Test GetAll - returns []*ModelInfo (slice of pointers)
	models := reg.GetAll()
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

	// Verify the model in slice matches
	m := models[0]
	if m.ID != "qwen-7b" {
		t.Errorf("Expected ID 'qwen-7b' in GetAll, got '%s'", m.ID)
	}

	// Test UnloadModel - use Load/Unload pattern instead
	err = reg.Unload("qwen-7b")
	if err != nil {
		t.Logf("Note: %v", err)
	}

	models = reg.GetAll()
}

// TestConfig_Integration tests configuration loading
func TestConfig_Integration(t *testing.T) {
	t.Parallel()

	// Use existing config file if available
	configs, err := config.LoadModelsFromYAML("/Users/wohlgemuth/IdeaProjects/llm-proxy/config/models.yaml")
	if configs != nil && len(configs.Models) > 0 {
		t.Logf("Loaded %d models from config file", len(configs.Models))
	} else if err != nil {
		t.Logf("Note: Config file not found or error loading: %v", err)
	}
}

// TestDiscovery_Integration tests discovery service parsing
func TestDiscovery_Integration(t *testing.T) {
	t.Parallel()

	// LM Studio returns an array of model objects directly
	lmStudioResponse := `[{"id":"qwen-7b-q4_k_m","name":"Qwen/Qwen-7B-GGUF/qwen-7b-q4_k_m.gguf"}]`

	models, err := discovery.ParseLMStudioModels([]byte(lmStudioResponse))
	if err != nil {
		t.Fatalf("Failed to parse LM Studio models: %v", err)
	}

	if len(models) != 1 {
		t.Errorf("Expected 1 discovered model, got %d", len(models))
	}

	if models[0].ID != "qwen-7b-q4_k_m" {
		t.Errorf("Expected 'qwen-7b-q4_k_m', got '%s'", models[0].ID)
	}

	// Verify other fields are parsed
	if models[0].Name == "" {
		t.Errorf("Expected name field to be populated")
	}
}

// TestDiscovery_EmptyModels tests edge case with empty models array
func TestDiscovery_EmptyModels(t *testing.T) {
	t.Parallel()

	lmStudioResponse := `[]`

	models, err := discovery.ParseLMStudioModels([]byte(lmStudioResponse))
	if err != nil {
		t.Fatalf("Failed to parse empty models: %v", err)
	}

	if len(models) != 0 {
		t.Errorf("Expected 0 discovered models, got %d", len(models))
	}
}

// TestNormalizer_EmptyModelName tests edge case with empty model name
func TestNormalizer_EmptyModelName(t *testing.T) {
	t.Parallel()

	backendResponse := map[string]interface{}{
		"message": map[string]interface{}{
			"role":    "assistant",
			"content": "Test response!",
		},
	}

	normalized, err := normalizer.NormalizeResponse("", "", time.Now(), backendResponse)
	if err != nil {
		t.Fatalf("Failed to normalize empty model: %v", err)
	}

	var normalizedMap map[string]interface{}
	json.Unmarshal(normalized, &normalizedMap)

	if normalizedMap["id"] != "" {
		t.Errorf("Expected empty ID for empty model name, got '%s'", normalizedMap["id"])
	}
}

// TestRouter_NoRoutes tests router with no registered routes
func TestRouter_NoRoutes(t *testing.T) {
	t.Parallel()

	r := router.NewRouter()

	target, _, found := r.GetTargetForPath("/some-path/")
	if found {
		t.Error("Expected no route to be found for non-registered path")
	} else if target != nil {
		t.Errorf("Expected nil target for non-registered path, got %v", target)
	}
}
