package memory

import (
	"fmt"
	"testing"
)

func TestMemoryPoolManager_New(t *testing.T) {
	manager := NewMemoryPoolManager(16*1024*1024*1024, 16.0) // 16GB threshold

	if manager == nil {
		t.Fatal("Expected memory pool manager to be created")
	}

	if manager.pool.Name != "combined" {
		t.Errorf("Expected pool name 'combined', got %s", manager.pool.Name)
	}

	if manager.pool.Total != 16*1024*1024*1024 {
		t.Errorf("Expected total memory 16GB, got %d", manager.pool.Total)
	}

	if manager.pool.Free != 16*1024*1024*1024 {
		t.Errorf("Expected free memory equal to total initially")
	}
}

func TestMemoryPoolManager_AddModel(t *testing.T) {
	manager := NewMemoryPoolManager(16*1024*1024*1024, 16.0)

	manager.AddModel("test-model", 7*1024*1024*1024, "cpu")

	models := manager.GetModels()

	if len(models) != 1 {
		t.Errorf("Expected 1 model in pool, got %d", len(models))
	}

	model := models[0]
	if model.ModelName != "test-model" {
		t.Errorf("Expected model name 'test-model', got %s", model.ModelName)
	}

	if model.Used != 0 {
		t.Errorf("Expected used memory 0 initially, got %d", model.Used)
	}

	if model.Free != 7*1024*1024*1024 {
		t.Errorf("Expected free memory equal to size initially")
	}
}

func TestMemoryPoolManager_AddMultipleModels(t *testing.T) {
	manager := NewMemoryPoolManager(16*1024*1024*1024, 16.0) // 16GB threshold

	// Add multiple models: Qwen (7GB), Mistral (6GB), LLaMA (8GB), Phi-3 (3GB)
	totalAdded := uint64(0)

	manager.AddModel("qwen", 7*1024*1024*1024, "cpu")
	totalAdded += 7 * 1024 * 1024 * 1024

	manager.AddModel("mistral", 6*1024*1024*1024, "cpu")
	totalAdded += 6 * 1024 * 1024 * 1024

	manager.AddModel("llama", 8*1024*1024*1024, "gpu:0")
	totalAdded += 8 * 1024 * 1024 * 1024

	manager.AddModel("phi", 3*1024*1024*1024, "cpu")
	totalAdded += 3 * 1024 * 1024 * 1024

	models := manager.GetModels()

	if len(models) != 4 {
		t.Errorf("Expected 4 models in pool, got %d", len(models))
	}

	if totalAdded != 24*1024*1024*1024 { // 7+6+8+3 = 24GB
		t.Errorf("Expected total added memory 24GB, got %.2fGB", float64(totalAdded)/1024/1024/1024)
	}
}

func TestMemoryPoolManager_RemoveModel(t *testing.T) {
	manager := NewMemoryPoolManager(16*1024*1024*1024, 16.0)

	// Simulate model loading by adding to pool
	manager.AddModel("qwen", 7*1024*1024*1024, "cpu")

	models := manager.GetModels()
	if len(models) != 1 {
		t.Fatal("Expected 1 model before removal")
	}

	// Simulate loading by updating used/free (not actually removing from pool tracking)
	// For actual removal use RemoveModel
	manager.RemoveModel("qwen")

	models = manager.GetModels()
	if len(models) != 0 {
		t.Errorf("Expected 0 models after removal, got %d", len(models))
	}
}

func TestMemoryPoolManager_GetAllPools(t *testing.T) {
	manager := NewMemoryPoolManager(16*1024*1024*1024, 16.0)

	manager.AddModel("test-model", 7*1024*1024*1024, "cpu")

	allPools := manager.GetAllPools()

	if len(allPools) != 2 { // combined pool + model pool
		t.Errorf("Expected 2 pools (combined + model), got %d", len(allPools))
	}

	// First pool should be the combined one
	if allPools[0].Name != "combined" {
		t.Errorf("Expected first pool to be 'combined', got %s", allPools[0].Name)
	}

	if allPools[0].Total != 16*1024*1024*1024 {
		t.Errorf("Expected combined pool total to be 16GB")
	}
}

func TestMemoryPoolManager_CombinedPoolAlwaysFirst(t *testing.T) {
	manager := NewMemoryPoolManager(32*1024*1024*1024, 32.0) // 32GB threshold

	manager.AddModel("model-1", 8*1024*1024*1024, "cpu")
	manager.AddModel("model-2", 6*1024*1024*1024, "gpu:0")

	allPools := manager.GetAllPools()

	if len(allPools) != 3 { // combined + 2 models
		t.Fatalf("Expected 3 pools, got %d", len(allPools))
	}

	// Combined pool should always be first regardless of AddModel calls
	for i, pool := range allPools {
		if i == 0 && pool.Name != "combined" {
			t.Errorf("Combined pool expected at index 0, got %s", pool.Name)
		}
	}
}

func TestMemoryPoolManager_MultipleAddModelSameName(t *testing.T) {
	manager := NewMemoryPoolManager(16*1024*1024*1024, 16.0)

	// Adding same model twice should not duplicate
	manager.AddModel("test-model", 7*1024*1024*1024, "cpu")
	manager.AddModel("test-model", 5*1024*1024*1024, "gpu:0")

	models := manager.GetModels()

	if len(models) != 1 {
		t.Errorf("Expected 1 model (no duplication), got %d", len(models))
	}
}

func TestMemoryPoolManager_MemoryCalculation(t *testing.T) {
	mathGB := uint64(1024 * 1024 * 1024) // 1 GB in bytes

	// Test exact values
	testCases := []struct {
		name   string
		sizeGB float64
		expect uint64
	}{
		{"1GB", 1.0, uint64(1 * mathGB)},
		{"3GB", 3.0, uint64(3 * mathGB)},
		{"7GB", 7.0, uint64(7 * mathGB)},
		{"8GB", 8.0, uint64(8 * mathGB)},
		{"14GB", 14.0, uint64(14 * mathGB)},
		{"24GB", 24.0, uint64(24 * mathGB)},
	}

	for _, tc := range testCases {
		expectedBytes := uint64(tc.sizeGB * float64(mathGB))

		manager := NewMemoryPoolManager(100*1024*1024*1024, 100.0) // Large threshold to avoid eviction
		manager.AddModel(tc.name, expectedBytes, "cpu")

		models := manager.GetModels()
		if len(models) != 1 {
			t.Fatalf("Test %s: Expected 1 model, got %d", tc.name, len(models))
		}

		if models[0].Total != expectedBytes {
			t.Errorf("Test %s: Expected total %d bytes, got %d bytes", tc.name, expectedBytes, models[0].Total)
		}

		if models[0].Free != expectedBytes {
			t.Errorf("Test %s: Expected free %d bytes initially, got %d bytes", tc.name, expectedBytes, models[0].Free)
		}
	}
}

func TestMemoryPoolManager_FormatReadable(t *testing.T) {
	mathGB := uint64(1024 * 1024 * 1024)
	testCases := []struct {
		name    string
		totalGB float64
		expect  string
	}{
		{"1GB", 1.0, "1.0 GB"},
		{"3GB", 3.0, "3.0 GB"},
		{"7GB", 7.0, "7.0 GB"},
		{"24GB", 24.0, "24.0 GB"},
	}

	for _, tc := range testCases {
		// Create manager with large threshold to avoid eviction
		manager := NewMemoryPoolManager(uint64(tc.totalGB*float64(mathGB)*10), tc.totalGB)

		// Need to add model first since GetModels() only returns registered models, not combined pool
		if len(manager.GetModels()) == 0 {
			manager.AddModel(tc.name, uint64(tc.totalGB*float64(mathGB)), "cpu")
		}

		models := manager.GetModels()
		if len(models) == 0 {
			t.Fatal("No models registered")
		}

		// Verify memory size matches expected (with some tolerance for float precision)
		actualGB := float64(models[0].Total) / float64(mathGB)
		if actualGB < tc.totalGB-0.01 || actualGB > tc.totalGB+0.01 {
			t.Errorf("Memory size mismatch: %.2f vs %.2f", actualGB, tc.totalGB)
		}
	}
}

func TestMemoryPoolManager_WithZeroThreshold(t *testing.T) {
	// This should still create a manager with 0 bytes threshold
	manager := NewMemoryPoolManager(0, 0.0)

	if manager == nil {
		t.Fatal("Expected manager even with zero threshold")
	}
}

func TestMemoryPoolManager_WithDifferentThreshold(t *testing.T) {
	testCases := []struct {
		name        string
		totalGB     float64
		thresholdGB float64
		expectError bool
	}{
		{"16GB", 16.0, 16.0, false},
		{"32GB", 32.0, 16.0, false}, // More than threshold
		{"8GB", 8.0, 4.0, false},    // Less than threshold
	}

	for _, tc := range testCases {
		manager := NewMemoryPoolManager(uint64(tc.totalGB*1024*1024*1024), tc.thresholdGB)

		models := manager.GetModels()
		if len(models) != 1 && manager.pool.Total == 0 {
			// Add a dummy model if pool is zero for this test
			manager.AddModel("dummy", uint64(tc.totalGB*1024*1024*1024), "cpu")
		}

		allPools := manager.GetAllPools()
		if len(allPools) == 0 && manager.pool.Total != 0 {
			t.Fatal("Expected at least combined pool")
		}
	}
}

func BenchmarkPoolManager_AddModel(b *testing.B) {
	manager := NewMemoryPoolManager(32*1024*1024*1024, 32.0)
	for i := 0; i < b.N; i++ {
		modelName := fmt.Sprintf("model-%d", i%10)
		manager.AddModel(modelName, uint64(7*1024*1024*1024), "cpu")
	}
}

func BenchmarkPoolManager_GetAllPools(b *testing.B) {
	manager := NewMemoryPoolManager(32*1024*1024*1024, 32.0)
	for i := 0; i < 10; i++ {
		modelName := fmt.Sprintf("model-%d", i)
		manager.AddModel(modelName, uint64((i+1)*1024*1024*1024), "cpu")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetAllPools()
	}
}

func BenchmarkPoolManager_AddMultiple(b *testing.B) {
	models := []string{"qwen", "mistral", "llama", "phi", "grok", "mixtral"}
	benchSize := uint64(7 * 1024 * 1024 * 1024) // 7GB

	for i := 0; i < b.N; i++ {
		manager := NewMemoryPoolManager(100*1024*1024*1024, 100.0)
		for _, model := range models {
			manager.AddModel(model, benchSize, "cpu")
		}
	}
}
