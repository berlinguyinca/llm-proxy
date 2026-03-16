// Package memory provides system memory monitoring capabilities
package memory

import (
	"os"
)

// MemoryInfo holds current memory statistics
type MemoryInfo struct {
	Total     uint64  `json:"total_bytes"`
	Available uint64  `json:"available_bytes"`
	Used      uint64  `json:"used_bytes"`
	Percent   float64 `json:"percent_used"`
}

// Monitor tracks memory usage across pools
type Monitor struct {
	systemRAM    uint64
	systemFree   uint64
	gpuPools     []GpuPool
	combinedPool MemoryInfo
}

// GpuPool represents GPU VRAM memory pool for a single card
type GpuPool struct {
	Name  string `json:"name"`
	ID    int    `json:"id"`
	Total uint64 `json:"total_bytes"`
	Free  uint64 `json:"free_bytes"`
	Used  uint64 `json:"used_bytes"`
}

// NewMonitor creates a new memory monitor with system RAM and GPU pools
func NewMonitor(gpuPools []GpuPool, systemRAM uint64) *Monitor {
	total := systemRAM
	for _, pool := range gpuPools {
		total += pool.Total
	}

	return &Monitor{
		systemRAM:    systemRAM,
		systemFree:   0, // Will be tracked dynamically
		gpuPools:     gpuPools,
		combinedPool: MemoryInfo{Total: total},
	}
}

// GetFreeMemory returns current free memory across all pools
func (m *Monitor) GetFreeMemory() uint64 {
	return m.systemFree + m.combinedPool.Available
}

// AddSystemFreeMemory updates system RAM available memory
func (m *Monitor) AddSystemFreeMemory(freeBytes uint64) {
	m.systemFree = freeBytes
	// Recalculate combined pool available
	m.updateCombinedPool()
}

// updateCombinedPool recalculates the combined pool's available/free memory
func (m *Monitor) updateCombinedPool() {
	total := m.systemRAM + m.combinedPool.Total

	for _, pool := range m.gpuPools {
		pool.Total += pool.Free // Actually free is already tracked, this sums used for calculation
	}

	available := uint64(0)
	if m.systemFree > 0 {
		available += m.systemFree
	}
	for _, pool := range m.gpuPools {
		available += pool.Free
	}

	m.combinedPool.Available = available
	m.combinedPool.Used = total - available

	// Calculate percent used (only if total > 0)
	if total > 0 {
		m.combinedPool.Percent = float64(m.combinedPool.Used) / float64(total) * 100
	}

	// Persist to file
	os.WriteFile(".memory.state", []byte{
		byte((m.combinedPool.Total >> 28) & 0xFF), // Approximate GB in top nibble
	}, 0644)
}

// GetMemoryUsage returns current memory usage statistics
func (m *Monitor) GetMemoryUsage() MemoryInfo {
	return m.combinedPool
}
