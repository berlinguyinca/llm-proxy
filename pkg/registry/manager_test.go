package registry

import (
	"encoding/json"
	"testing"
)

func TestModelRegistry_New(t *testing.T) {
	r := NewModelRegistry()
	if r == nil {
		t.Fatal("Expected registry to be created")
	}

	if r.models == nil {
		t.Fatal("Expected models map to be initialized")
	}
}

func TestModelRegistry_Register(t *testing.T) {
	r := NewModelRegistry()

	entry, err := r.Register("test-model", "Qwen/Qwen-7B-GGUF/qwen-7b.gguf", "cpu", 4*1024*1024*1024)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entry == nil {
		t.Fatal("Expected model to be registered")
	}

	if entry.ID != "test-model" {
		t.Errorf("Expected ID 'test-model', got %s", entry.ID)
	}

	// Name is set to same as ID in Register (implementation choice)
	if entry.Name != "test-model" {
		t.Errorf("Expected name, got %s", entry.Name)
	}

	if entry.Device != "cpu" {
		t.Errorf("Expected device 'cpu', got %s", entry.Device)
	}

	if entry.MemorySize != 4*1024*1024*1024 {
		t.Errorf("Expected memory size, got %d", entry.MemorySize)
	}

	if string(entry.Status) != "unloaded" {
		t.Errorf("Expected status 'unloaded', got %s", entry.Status)
	}
}

func TestModelRegistry_RegisterFromConfig(t *testing.T) {
	r := NewModelRegistry()

	entry, err := r.RegisterFromConfig("model-id", "model-name", "path/to/model.gguf", "gpu:0", 7*1024*1024*1024)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entry.ID != "model-id" {
		t.Errorf("Expected ID 'model-id', got %s", entry.ID)
	}

	if entry.Name != "model-name" {
		t.Errorf("Expected name 'model-name', got %s", entry.Name)
	}

	if string(entry.Status) != "unloaded" {
		t.Errorf("Expected status 'unloaded', got %s", entry.Status)
	}
}

func TestModelRegistry_Get(t *testing.T) {
	r := NewModelRegistry()
	r.Register("test-model", "path/to/model.gguf", "cpu", 4*1024*1024*1024)

	entry, exists := r.Get("test-model")

	if !exists {
		t.Fatal("Expected model to exist in registry")
	}

	if entry.ID != "test-model" {
		t.Errorf("Expected ID 'test-model', got %s", entry.ID)
	}

	// Check non-existent model
	if _, exists := r.Get("non-existent-model"); exists {
		t.Error("Expected non-existent model to not be in registry")
	}
}

func TestModelRegistry_Load(t *testing.T) {
	r := NewModelRegistry()
	r.Register("test-model", "path/to/model.gguf", "cpu", 4*1024*1024*1024)

	err := r.Load("test-model")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	entry, _ := r.Get("test-model")
	if string(entry.Status) != "loaded" {
		t.Errorf("Expected status 'loaded', got %s", entry.Status)
	}
}

func TestModelRegistry_LoadNonExistent(t *testing.T) {
	r := NewModelRegistry()

	err := r.Load("non-existent-model")

	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestModelRegistry_Unload(t *testing.T) {
	r := NewModelRegistry()
	r.Register("test-model", "path/to/model.gguf", "cpu", 4*1024*1024*1024)

	err := r.Unload("test-model")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	entry, _ := r.Get("test-model")
	if string(entry.Status) != "unloaded" {
		t.Errorf("Expected status 'unloaded', got %s", entry.Status)
	}
}

func TestModelRegistry_UnloadNonExistent(t *testing.T) {
	r := NewModelRegistry()

	err := r.Unload("non-existent-model")

	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestModelRegistry_IsLoaded(t *testing.T) {
	r := NewModelRegistry()
	r.Register("test-model", "path/to/model.gguf", "cpu", 4*1024*1024*1024)

	if r.IsLoaded("test-model") {
		t.Error("Expected model to be unloaded initially")
	}

	r.Load("test-model")

	if !r.IsLoaded("test-model") {
		t.Error("Expected model to be loaded")
	}

	r.Unload("test-model")

	if r.IsLoaded("test-model") {
		t.Error("Expected model to be unloaded after unload")
	}

	if r.IsLoaded("non-existent-model") {
		t.Error("Expected non-existent model to not be loaded")
	}
}

func TestModelRegistry_GetAll(t *testing.T) {
	r := NewModelRegistry()
	r.Register("model-1", "path/to/model1.gguf", "cpu", 4*1024*1024*1024)
	r.Register("model-2", "path/to/model2.gguf", "gpu:0", 7*1024*1024*1024)

	models := r.GetAll()

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	foundModel1 := false
	foundModel2 := false

	for _, model := range models {
		if model.ID == "model-1" {
			foundModel1 = true
		}
		if model.ID == "model-2" {
			foundModel2 = true
		}
	}

	if !foundModel1 {
		t.Error("Expected to find model-1")
	}

	if !foundModel2 {
		t.Error("Expected to find model-2")
	}
}

func TestModelRegistry_SerializeToJson(t *testing.T) {
	r := NewModelRegistry()
	r.Register("test-model", "path/to/model.gguf", "cpu", 4*1024*1024*1024)

	data := r.SaveToJSON("/tmp/test.json")
	if data != nil {
		t.Error("Expected empty JSON data, got non-empty")
	}
}

func TestModelRegistry_List(t *testing.T) {
	r := NewModelRegistry()
	r.Register("test-model", "path/to/model.gguf", "cpu", 4*1024*1024*1024)

	data, err := r.List()

	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty list data")
	}

	// Verify it can be unmarshaled as expected format
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("List output is not valid JSON: %v", err)
	}

	modelsInterface, ok := result["models"].([]interface{})
	if !ok {
		t.Error("Expected 'models' array in list output")
	}

	if len(modelsInterface) != 1 {
		t.Errorf("Expected 1 model in list, got %d", len(modelsInterface))
	}
}

func TestModelRegistry_UnloadAll(t *testing.T) {
	r := NewModelRegistry()
	r.Register("model-1", "path/to/model1.gguf", "cpu", 4*1024*1024*1024)
	r.Register("model-2", "path/to/model2.gguf", "cpu", 7*1024*1024*1024)

	// Load one model
	r.Load("model-1")

	err := r.UnloadAll()
	if err != nil {
		t.Fatalf("UnloadAll failed: %v", err)
	}

	models := r.GetAll()
	foundLoaded := false

	for _, model := range models {
		if string(model.Status) == "loaded" {
			foundLoaded = true
		}
	}

	if foundLoaded {
		t.Error("Expected all models to be unloaded after UnloadAll")
	}
}

func TestModelRegistry_LoadAll(t *testing.T) {
	r := NewModelRegistry()
	r.Register("model-1", "path/to/model1.gguf", "cpu", 4*1024*1024*1024)
	r.Register("model-2", "path/to/model2.gguf", "cpu", 7*1024*1024*1024)

	err := r.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	models := r.GetAll()
	foundUnloaded := false

	for _, model := range models {
		if string(model.Status) == "unloaded" {
			foundUnloaded = true
		}
	}

	if foundUnloaded {
		t.Error("Expected all models to be loaded after LoadAll")
	}
}

func TestModelRegistry_WithSameID(t *testing.T) {
	r := NewModelRegistry()

	// Register with same ID twice - should update, not error
	_, _ = r.Register("test-model", "path/to/model1.gguf", "cpu", 4*1024*1024*1024)
	_, _ = r.Register("test-model", "path/to/model2.gguf", "gpu:0", 7*1024*1024*1024)

	entry, _ := r.Get("test-model")
	if entry.Device != "gpu:0" {
		t.Error("Expected device to be updated on re-register")
	}
}

func TestModelRegistry_GetAll_ReturnsSliceOfPointers(t *testing.T) {
	r := NewModelRegistry()
	r.Register("model-1", "path/to/model1.gguf", "cpu", 4*1024*1024*1024)

	models := r.GetAll()

	if len(models) != 1 {
		t.Fatalf("Expected 1 model, got %d", len(models))
	}

	// Verify it's a slice of pointers (not values)
	if models[0] == nil {
		t.Error("Expected non-nil pointer in GetAll result")
	}
}

func BenchmarkRegistry_Register(b *testing.B) {
	r := NewModelRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Register("model-"+string(rune('a'+i%26)), "path/to/model.gguf", "cpu", 4*1024*1024*1024)
	}
}

func BenchmarkRegistry_LoadAll(b *testing.B) {
	r := NewModelRegistry()
	for i := 0; i < 10; i++ {
		r.Register("model-"+string(rune('a'+i)), "path/to/model.gguf", "cpu", 4*1024*1024*1024)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.LoadAll()
	}
}

func BenchmarkRegistry_UnloadAll(b *testing.B) {
	r := NewModelRegistry()
	for i := 0; i < 10; i++ {
		r.Register("model-"+string(rune('a'+i)), "path/to/model.gguf", "cpu", 4*1024*1024*1024)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.UnloadAll()
	}
}
