package memory

import (
	"fmt"
)

// MemoryPool represents a memory pool for tracking usage
type MemoryPool struct {
	Name      string `json:"name"`
	Total     uint64 `json:"total_bytes"`
	Used      uint64 `json:"used_bytes"`
	Free      uint64 `json:"free_bytes"`
	ModelName string `json:"model_name"`
}

// MemoryPoolManager manages memory pools for models
type MemoryPoolManager struct {
	pool   MemoryPool
	models map[string]*MemoryPool
}

// NewMemoryPoolManager creates a new manager with combined pool
func NewMemoryPoolManager(totalMemory uint64, _ float64) *MemoryPoolManager {
	return &MemoryPoolManager{
		pool: MemoryPool{
			Name:  "combined",
			Total: totalMemory,
			Used:  0,
			Free:  totalMemory,
		},
		models: make(map[string]*MemoryPool),
	}
}

// AddModel adds a model to the memory pool tracking
func (m *MemoryPoolManager) AddModel(modelName string, sizeBytes uint64, deviceType string) {
	if m.models[modelName] != nil {
		return // Already exists
	}

	m.models[modelName] = &MemoryPool{
		Name:      modelName,
		Total:     sizeBytes,
		Used:      0,
		Free:      sizeBytes,
		ModelName: modelName,
	}

	return
}

// RemoveModel removes a model from the memory pool tracking
func (m *MemoryPoolManager) RemoveModel(modelName string) {
	delete(m.models, modelName)
	fmt.Printf("Removed model %s from memory pool\n", modelName)
}

// GetModels returns all tracked models
func (m *MemoryPoolManager) GetModels() []*MemoryPool {
	models := make([]*MemoryPool, 0, len(m.models))
	for _, pool := range m.models {
		models = append(models, pool)
	}
	return models
}

// GetAllPools returns all pools including the combined one
func (m *MemoryPoolManager) GetAllPools() []*MemoryPool {
	pools := make([]*MemoryPool, 0, len(m.models)+1)

	// Add combined pool first
	pools = append(pools, &m.pool)

	// Then add model pools
	for _, pool := range m.models {
		pools = append(pools, pool)
	}

	return pools
}
