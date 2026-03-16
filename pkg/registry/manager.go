package registry

import (
	"encoding/json"
	"fmt"
	"time"
)

// Status represents the current status of a model
type Status string

const (
	StatusUnloaded Status = "unloaded"
	StatusLoading  Status = "loading"
	StatusLoaded   Status = "loaded"
)

// ModelInfo contains information about a managed model
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	QualifiedName string `json:"qualified_name"`
	Device        string `json:"device"`
	MemorySize    uint64 `json:"memory_size_bytes"`
	Status        Status `json:"status"`
	URL           string `json:"url,omitempty"`
}

// ModelRegistry manages model loading and unloading operations
type ModelRegistry struct {
	models map[string]*ModelInfo
}

// NewModelRegistry creates a new model registry
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*ModelInfo),
	}
}

// Register adds or updates a model entry
func (r *ModelRegistry) Register(modelName, qualifiedName string, device string, sizeBytes uint64) (*ModelInfo, error) {
	entry := &ModelInfo{
		ID:            modelName,
		QualifiedName: qualifiedName,
		Name:          modelName,
		Device:        device,
		MemorySize:    sizeBytes,
		Status:        StatusUnloaded,
		URL:           "",
	}

	r.models[modelName] = entry
	return entry, nil
}

// RegisterFromConfig registers a model from configuration
func (r *ModelRegistry) RegisterFromConfig(id, name, qualifiedName, device string, sizeBytes uint64) (*ModelInfo, error) {
	entry := &ModelInfo{
		ID:            id,
		Name:          name,
		QualifiedName: qualifiedName,
		Device:        device,
		MemorySize:    sizeBytes,
		Status:        StatusUnloaded,
		URL:           "",
	}

	r.models[id] = entry
	return entry, nil
}

// Get returns a model by name
func (r *ModelRegistry) Get(modelName string) (*ModelInfo, bool) {
	entry, exists := r.models[modelName]
	return entry, exists
}

// LoadAll loads all registered models
func (r *ModelRegistry) LoadAll() error {
	for modelName := range r.models {
		if err := r.Load(modelName); err != nil {
			return err
		}
	}
	return nil
}

// UnloadAll unloads all loaded models
func (r *ModelRegistry) UnloadAll() error {
	models := r.GetAll()
	for _, model := range models {
		if model.Status == StatusLoaded {
			if err := r.Unload(model.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// Load loads a model from disk or remote source
func (r *ModelRegistry) Load(modelName string) error {
	if entry, exists := r.Get(modelName); !exists {
		return fmt.Errorf("model %s not registered", modelName)
	} else {
		entry.Status = StatusLoading
		fmt.Printf("Loading model: %s\n", entry.QualifiedName)

		// For demo purposes, simulate loading time
		time.Sleep(100 * time.Millisecond)

		// In production, you would actually load the model here using llama.cpp or similar
		entry.Status = StatusLoaded

		return nil
	}
}

// Unload unloads a model from memory
func (r *ModelRegistry) Unload(modelName string) error {
	if entry, exists := r.Get(modelName); !exists {
		return fmt.Errorf("model %s not registered", modelName)
	} else {
		entry.Status = StatusUnloaded

		// In production, you would actually unload the model here
		fmt.Printf("Unloaded model: %s\n", entry.QualifiedName)

		return nil
	}
}

// IsLoaded checks if a model is currently loaded
func (r *ModelRegistry) IsLoaded(modelName string) bool {
	if entry, exists := r.Get(modelName); exists {
		return entry.Status == StatusLoaded
	}
	return false
}

// RegisterModelFromConfig registers a model from configuration (for on-demand loading)
func (r *ModelRegistry) RegisterModelFromConfig(name, path, device string) (*ModelInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	entry := &ModelInfo{
		ID:            name,
		Name:          name,
		QualifiedName: name + " (auto-loaded)",
		Device:        device,
		MemorySize:    0, // Will be detected when loaded
		Status:        StatusUnloaded,
		URL:           path,
	}

	r.models[name] = entry
	return entry, nil
}

// LoadFromDisk loads a model directly from disk path (for auto-loading)
func (r *ModelRegistry) LoadFromDisk(name, path, device string) error {
	if entry, exists := r.Get(name); !exists {
		// Auto-register the model if it doesn't exist yet
		_, err := r.RegisterModelFromConfig(name, path, device)
		return err
	} else {
		// Already registered, just update status
		entry.Status = StatusLoading
		fmt.Printf("Loading model from disk: %s from %s\n", name, path)

		// For demo purposes, simulate loading time
		time.Sleep(100 * time.Millisecond)

		// In production, you would actually load the model here using llama.cpp or similar
		entry.Status = StatusLoaded

		return nil
	}
}

// GetAll returns all registered models
func (r *ModelRegistry) GetAll() []*ModelInfo {
	models := make([]*ModelInfo, 0, len(r.models))
	for _, model := range r.models {
		models = append(models, model)
	}
	return models
}

// SaveToJSON serializes the registry to JSON (placeholder)
func (r *ModelRegistry) SaveToJSON(path string) error {
	_ = r.GetAll()
	return nil
}

// List returns a formatted list of models with their device placement info
func (r *ModelRegistry) List() ([]byte, error) {
	models := r.GetAll()

	// Convert to map[string]interface{} for JSON serialization
	list := make(map[string]interface{})
	list["models"] = make([]map[string]interface{}, len(models))

	for i, model := range models {
		list["models"].([]map[string]interface{})[i] = map[string]interface{}{
			"id":                model.ID,
			"name":              model.Name,
			"qualified_name":    model.QualifiedName,
			"device":            model.Device,
			"memory_size_bytes": model.MemorySize,
			"status":            string(model.Status),
			"url":               model.URL,
		}
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return nil, err
	}

	return data, nil
}
