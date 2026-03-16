package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// ModelInfo represents a model from LM Studio's /api/v1/models endpoint
type ModelInfo struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	SizeInBytes  int64      `json:"size_in_bytes"`
	Digest       string     `json:"digest"`
	ModelFamily  string     `json:"model_family"`
	Format       string     `json:"format"`
	Quantization string     `json:"quantization"`
	State        ModelState `json:"state"`
	CreatedAt    *string    `json:"created_at"`
	ModifiedAt   *string    `json:"modified_at"`
}

// ModelState represents whether a model is loaded or not
type ModelState struct {
	Loaded bool   `json:"loaded"`
	Path   string `json:"path,omitempty"`
}

// LmStudioDiscovery holds configuration for LM Studio discovery
type LmStudioDiscovery struct {
	URL     string `yaml:"url"`
	Regex   string `yaml:"regex"`
	Enabled bool   `yaml:"enabled"`
}

var defaultRegex = `(?<model>[^/]+)\.gguf`

// ExtractModelName extracts the model name from a LM Studio model path
func ExtractModelName(path string) (string, error) {
	re := regexp.MustCompile(defaultRegex)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract model name from path: %s", path)
	}
	return matches[1], nil
}

// ParseLMStudioModels parses the /api/v1/models endpoint response
func ParseLMStudioModels(response []byte) ([]ModelInfo, error) {
	var models []ModelInfo

	if err := json.Unmarshal(response, &models); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LM Studio models: %w", err)
	}

	return models, nil
}

// DiscoverModels fetches and parses models from LM Studio
func (d *LmStudioDiscovery) DiscoverModels() ([]ModelInfo, error) {
	if !d.Enabled || d.URL == "" {
		return nil, fmt.Errorf("discovery disabled or URL not configured")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", d.URL, nil)
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

	models, err := ParseLMStudioModels(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse models: %w", err)
	}

	return models, nil
}
