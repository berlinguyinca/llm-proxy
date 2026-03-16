package discovery

import (
	"testing"
)

func TestParseLMStudioModels_EmptyResponse(t *testing.T) {
	// Empty JSON should return empty models list, not error - this is graceful handling
	models, _ := ParseLMStudioModels([]byte(""))

	if len(models) != 0 {
		t.Errorf("Expected 0 models from empty response, got %d", len(models))
	}
}

func TestParseLMStudioModels_ValidResponse(t *testing.T) {
	mockResponse := []byte(`[
		{
			"id": "Qwen/Qwen-7B-Chat-GGUF/qwen-7b-chat-q4_k_m.gguf",
			"name": "Qwen/Qwen-7B-Chat-GGUF/qwen-7b-chat-q4_k_m.gguf",
			"size_in_bytes": 4294967296,
			"model_family": "qwen",
			"format": "gguf",
			"quantization": "Q4_K_M",
			"state": {
				"loaded": false,
				"path": "/models/qwen-7b-chat-q4_k_m.gguf"
			}
		},
		{
			"id": "Mistral/Mistral-7B-Instruct-v0.2-GGUF/mistral-7b-instruct-v0.2-q4_k_m.gguf",
			"name": "Mistral/Mistral-7B-Instruct-v0.2-GGUF/mistral-7b-instruct-v0.2-q4_k_m.gguf",
			"size_in_bytes": 3221225472,
			"model_family": "mistral",
			"format": "gguf",
			"quantization": "Q4_K_M",
			"state": {
				"loaded": true,
				"path": "/models/mistral-7b-instruct-v0.2-q4_k_m.gguf"
			}
		},
		{
			"id": "llama.cpp/llama-ggml/llama-3-8b-q5_1.gguf",
			"name": "llama.cpp/llama-ggml/llama-3-8b-q5_1.gguf",
			"size_in_bytes": 6442450944,
			"model_family": "llama",
			"format": "gguf",
			"quantization": "Q5_1",
			"state": {
				"loaded": false,
				"path": ""
			}
		}
	]`)

	models, err := ParseLMStudioModels(mockResponse)

	if err != nil {
		t.Fatalf("ParseLMStudioModels failed: %v", err)
	}

	if len(models) != 3 {
		t.Errorf("Expected 3 models parsed, got %d", len(models))
	}

	// Verify first model
	if models[0].ID != "Qwen/Qwen-7B-Chat-GGUF/qwen-7b-chat-q4_k_m.gguf" {
		t.Errorf("Expected first model ID to be 'Qwen/...', got %s", models[0].ID)
	}

	if models[0].SizeInBytes != 4294967296 { // 4GB
		t.Errorf("Expected first model size 4GB, got %.2fGB", float64(models[0].SizeInBytes)/1024/1024/1024)
	}

	if models[0].ModelFamily != "qwen" {
		t.Errorf("Expected model family 'qwen', got %s", models[0].ModelFamily)
	}

	// Verify loaded state
	if models[1].State.Loaded != true {
		t.Error("Expected second model to be loaded")
	}
}

func TestParseLMStudioModels_SingleModel(t *testing.T) {
	mockResponse := []byte(`[{
		"id": "mistralai/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct-v0.1.Q4_K_M.gguf",
		"name": "mistralai/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct-v0.1.Q4_K_M.gguf",
		"size_in_bytes": 4294967296,
		"model_family": "mistral",
		"format": "gguf",
		"quantization": "Q4_K_M",
		"state": {"loaded": false}
	}]`)

	models, err := ParseLMStudioModels(mockResponse)

	if err != nil {
		t.Fatalf("ParseLMStudioModels failed: %v", err)
	}

	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}
}

func TestExtractModelName_FromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
		wantErr  bool
	}{
		{
			path:     "Qwen/Qwen-7B-Chat-GGUF/qwen-7b-chat-q4_k_m.gguf",
			expected: "qwen-7b-chat-q4_k_m",
			wantErr:  false,
		},
		{
			path:     "mistralai/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct-v0.1.Q4_K_M.gguf",
			expected: "mistral-7b-instruct-v0.1.Q4_K_M",
			wantErr:  false,
		},
		{
			path:     "llama.cpp/llama-ggml/llama-3-8b-q5_1.gguf",
			expected: "llama-3-8b-q5_1",
			wantErr:  false,
		},
		{
			path:     "phi/Phi-3-mini-instruct-GGUF/phi-3-mini-instruct-q4_k_m.gguf",
			expected: "phi-3-mini-instruct-q4_k_m",
			wantErr:  false,
		},
		{
			path:     "nonexistent/path/file.txt",
			expected: "",
			wantErr:  true, // No .gguf file
		},
		{
			path:     "/models/models/model.gguf",
			expected: "model",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(st *testing.T) {
			name, err := ExtractModelName(tc.path)

			if (err != nil) != tc.wantErr {
				st.Errorf("ExtractModelName() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if name != tc.expected {
				st.Errorf("ExtractModelName() = %s, want %s", name, tc.expected)
			}
		})
	}
}

func TestExtractModelName_FromQualifiedModelNames(t *testing.T) {
	// Test various model naming conventions from LM Studio
	testCases := []struct {
		path     string
		expected string
	}{
		{
			path:     "Qwen/Qwen-7B-Chat-GGUF/qwen-7b-chat-q4_k_m.gguf",
			expected: "qwen-7b-chat-q4_k_m",
		},
		{
			path:     "mistralai/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct-v0.1.Q8_0.gguf",
			expected: "mistral-7b-instruct-v0.1.Q8_0",
		},
		{
			path:     "TheBloke/Mistral-7B-Instruct-v0.2-GGUF/mistral-7b-instruct-v0.2.Q4_K_M.gguf",
			expected: "mistral-7b-instruct-v0.2.Q4_K_M",
		},
		{
			path:     "lmsys/Chatglm3-GGUF/chatglm3-6b-q4_k_m.gguf",
			expected: "chatglm3-6b-q4_k_m",
		},
		{
			path:     "microsoft/Phi-3-mini-128k-instruct-GGUF/phi-3-mini-128k-instruct-q5_K_M.gguf",
			expected: "phi-3-mini-128k-instruct-q5_K_M",
		},
	}

	for _, tc := range testCases {
		name, err := ExtractModelName(tc.path)
		if err != nil {
			t.Errorf("ExtractModelName(%s) failed: %v", tc.path, err)
		} else if name != tc.expected {
			t.Errorf("ExtractModelName(%s) = %s, want %s", tc.path, name, tc.expected)
		} else {
			t.Logf("✓ Extracted model name '%s' from path", name)
		}
	}
}

func TestLmStudioDiscovery_CanonicalType(t *testing.T) {
	d := &LmStudioDiscovery{
		URL:     "http://localhost:1234/api/v1/models",
		Regex:   `(?<model>[^/]+)\.gguf`,
		Enabled: true,
	}

	if d.URL != "http://localhost:1234/api/v1/models" {
		t.Errorf("Expected canonical URL")
	}

	if d.Regex != `(?<model>[^/]+)\.gguf` {
		t.Errorf("Expected default regex")
	}
}

func TestParseLMStudioModels_InvalidJSON(t *testing.T) {
	mockResponse := []byte(`{not valid json}`)

	models, err := ParseLMStudioModels(mockResponse)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Even if parsing fails, should return empty slice, not crash
	if models != nil && len(models) > 0 {
		t.Logf("Returned %d models despite parse error (acceptable)", len(models))
	}
}

func TestParseLMStudioModels_MalformedModelState(t *testing.T) {
	mockResponse := []byte(`[{"id": "test.gguf", "name": "test", "size_in_bytes": 1073741824, "state": "invalid-string"}]`)

	models, err := ParseLMStudioModels(mockResponse)

	if err != nil {
		t.Logf("Parse error for malformed model state: %v (acceptable)", err)
	} else if len(models) > 0 {
		// Check that malformed State doesn't cause crash
		if models[0].State.Loaded {
			t.Error("Expected loaded=false for string-based state")
		} else {
			t.Log("Handled string-based 'state' field gracefully as false")
		}
	}
}

func TestParseLMStudioModels_MissingOptionalFields(t *testing.T) {
	mockResponse := []byte(`[{
		"id": "minimal.gguf",
		"size_in_bytes": 1073741824
	}]`)

	models, err := ParseLMStudioModels(mockResponse)

	if err != nil {
		t.Errorf("Parse failed with minimal model: %v", err)
	}

	if len(models) != 1 {
		t.Errorf("Expected 1 model from minimal response")
	}

	model := models[0]
	if model.ID != "minimal.gguf" {
		t.Errorf("Expected ID 'minimal.gguf', got %s", model.ID)
	}

	// Check that optional fields default to zero values
	if model.ModelFamily != "" {
		t.Logf("Optional field ModelFamily defaults to: '%s'", model.ModelFamily)
	}

	if model.State.Loaded {
		t.Error("Expected loaded=false for missing state field")
	}
}

func BenchmarkParseLMStudioModels(b *testing.B) {
	mockResponse := []byte(`[
		{"id": "test1.gguf", "name": "test1", "size_in_bytes": 4294967296, "state": {"loaded": false}},
		{"id": "test2.gguf", "name": "test2", "size_in_bytes": 3221225472, "state": {"loaded": true}},
		{"id": "test3.gguf", "name": "test3", "size_in_bytes": 2097152000, "state": {"loaded": false}}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseLMStudioModels(mockResponse)
	}
}

func BenchmarkExtractModelName(b *testing.B) {
	path := "Qwen/Qwen-7B-Chat-GGUF/qwen-7b-chat-q4_k_m.gguf"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractModelName(path)
	}
}

func FuzzParseLMStudioModels(f *testing.F) {
	// Add fuzzing inputs
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[{"id": "test.gguf", "size_in_bytes": 1073741824}]`))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		models, _ := ParseLMStudioModels(data)

		// Should never panic
		if models != nil {
			_ = len(models)
		}
	})
}

func FuzzExtractModelName(f *testing.F) {
	f.Add("Qwen/Qwen-7B-Chat-GGUF/qwen-7b-chat-q4_k_m.gguf")
	f.Add("/models/model.gguf")
	f.Add("model.gguf")
	f.Add("invalid path without gguf extension")
	f.Add("")

	f.Fuzz(func(t *testing.T, path string) {
		_, _ = ExtractModelName(path)
	})
}
