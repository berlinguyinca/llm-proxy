package config

import (
	"os"
	"testing"
)

func TestLoadEnvironmentVars(t *testing.T) {
	tests := []struct {
		name            string
		envPrefix       string
		envVars         map[string]string
		expectedEnv     map[string]string
	}{
		{
			name:       "Load OpenAI API key",
			envPrefix:  "LLM_API_KEY_",
			envVars:    map[string]string{"LLM_API_KEY_OPENAI_API_KEY": "sk-test-openai-key"},
			expectedEnv: map[string]string{"OPENAI": "sk-test-openai-key"},
		},
		{
			name:       "Load Anthropic API key",
			envPrefix:  "LLM_API_KEY_",
			envVars:    map[string]string{"LLM_API_KEY_ANTHROPIC_API_KEY": "sk-test-anthropic-key"},
			expectedEnv: map[string]string{"ANTHROPIC": "sk-test-anthropic-key"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment variables for this test
			for key, value := range tc.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set env var %s: %v", key, err)
				}
			}

			defer func(keys []string) {
				for _, key := range keys {
					os.Unsetenv(key)
				}
			}(make([]string, 0, len(tc.envVars)))

			loader := NewConfigLoader("", tc.envPrefix)
			env := loader.LoadEnvironmentVars()

			for expectedKey, expectedValue := range tc.expectedEnv {
				if actualValue, ok := env[expectedKey]; !ok {
					t.Errorf("Expected key '%s' not found in loaded environment", expectedKey)
				} else if actualValue != expectedValue {
					t.Errorf("Expected value '%s' for key '%s', got '%s'", expectedValue, expectedKey, actualValue)
				}
			}
		})
	}
}

func TestLoadYAML(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		expectError    bool
	}{
		{
			name: "Valid YAML with models",
			yamlContent: `models:
  - id: qwen-7b
    name: Qwen-7B-Chat-GGUF
    url: http://localhost:1234/v1/chat/completions
    size_gb: 7.0
    device: cpu`,
			expectError: false,
		},
		{
			name: "Empty models array",
			yamlContent: `models: []`,
			expectError: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := t.TempDir() + "/test.yaml"
			if err := os.WriteFile(tmpFile, []byte(tc.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			loader := NewConfigLoader(tmpFile, "")
			configs, err := loader.LoadYAML()

			if tc.expectError && err == nil {
				t.Error("Expected error loading YAML")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if configs == nil && len(tc.yamlContent) > 0 {
				t.Error("Expected non-nil configs for valid YAML")
			}
		})
	}
}

func TestGetEnvironmentPrefix(t *testing.T) {
	tests := []struct {
		name      string
		envPrefix string
		expected  string
	}{
		{"Empty prefix", "", ""},
		{"Non-empty prefix", "LLM_API_KEY_", "LLM_API_KEY__"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			loader := NewConfigLoader("/tmp/test.yaml", tc.envPrefix)
			result := loader.GetEnvironmentPrefix()

			if result != tc.expected {
				t.Errorf("Expected prefix '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
