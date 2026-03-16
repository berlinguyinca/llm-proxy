// Package config provides configuration loading utilities
package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// ModelConfig represents a model's configuration from models.yaml
type ModelConfig struct {
	ID            string  `yaml:"id"`
	Name          string  `yaml:"name"`
	URL           string  `yaml:"url"`
	SizeGB        float64 `yaml:"size_gb"`
	Device        string  `yaml:"device"` // cpu or gpu:N
	QualifiedName string  `yaml:"qualified_name"`

	// Auto-load resource hints (optional, used by auto-load feature)
	min_memory_mb     int  `yaml:"min_memory_mb"`     // Minimum RAM required when model is loaded
	vram_mb_hint      int  `yaml:"vram_mb_hint"`      // Suggested VRAM hint for GPU placement (optional)
	eviction_priority int  `yaml:"eviction_priority"` // Higher = evicted first during memory pressure (1-10 scale)
	discovery_enabled bool `yaml:"discovery_enabled"` // Enable auto-discovery of this model from LM Studio
}

// ModelConfigs holds all model configurations from YAML file
type ModelConfigs struct {
	Models []ModelConfig `yaml:"models"`
}

// ConfigLoader loads configuration from multiple sources
type ConfigLoader struct {
	yamlPath  string
	envPrefix string
}

// NewConfigLoader creates a new config loader
func NewConfigLoader(yamlPath, envPrefix string) *ConfigLoader {
	return &ConfigLoader{
		yamlPath:  yamlPath,
		envPrefix: envPrefix,
	}
}

// LoadEnvironmentVars loads API keys from environment variables
func (c *ConfigLoader) LoadEnvironmentVars() map[string]string {
	env := make(map[string]string)

	if apiKey, ok := os.LookupEnv(c.envPrefix + "OPENAI_API_KEY"); ok {
		env["OPENAI"] = apiKey
	}
	if apiKey, ok := os.LookupEnv(c.envPrefix + "ANTHROPIC_API_KEY"); ok {
		env["ANTHROPIC"] = apiKey
	}

	return env
}

// LoadYAML loads model configurations from YAML file
func (c *ConfigLoader) LoadYAML() (*ModelConfigs, error) {
	data, err := os.ReadFile(c.yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configs ModelConfigs
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &configs, nil
}

// Load combines environment variables and YAML configurations
func (c *ConfigLoader) Load() (*ModelConfigs, error) {
	configs, err := c.LoadYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to load YAML config: %w", err)
	}

	return configs, nil
}

// LoadModelsFromYAML loads model configurations from YAML file (exported for main package)
func LoadModelsFromYAML(path string) (*ModelConfigs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configs ModelConfigs
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &configs, nil
}

// GetEnvironmentPrefix returns the environment variable prefix
func (c *ConfigLoader) GetEnvironmentPrefix() string {
	if c.envPrefix == "" {
		return ""
	}
	return c.envPrefix + "_"
}
