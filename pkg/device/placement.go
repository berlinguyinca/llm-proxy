// Package device provides device placement decision logic
package device

import (
	"math"
)

// DeviceType represents the type of compute device
type DeviceType string

const (
	GPU                DeviceType = "gpu"
	CPU                DeviceType = "cpu"
	INFERGPUDeviceType            = "infer_gpu" // Integrated GPU / AMD/Intel
)

// ModelPlacement holds a model and its current placement info
type ModelPlacement struct {
	ModelName  string     `json:"model_name"`
	DeviceType DeviceType `json:"device_type"`
	GPUIndex   int        `json:"gpu_index,omitempty"`
	MemorySize uint64     `json:"memory_size_bytes"`
	DeviceName string     `json:"device_name,omitempty"`
}

// GpuPoolInfo holds information about a GPU pool for placement decisions
type GpuPoolInfo struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	Capacity uint64 `json:"capacity_bytes"`
	Used     uint64 `json:"used_bytes"`
	Free     uint64 `json:"free_bytes"`
}

// PlacementDecision represents a device placement decision with reasoning
type PlacementDecision struct {
	DeviceType  DeviceType `json:"device_type"`
	GPUIndex    int        `json:"gpu_index,omitempty"`
	Reason      string     `json:"reason"`
	CanFitOnGPU bool       `json:"can_fit_on_gpu"`
}

// DecisionContext provides context for placement decisions
type DecisionContext struct {
	gpuPools    []GpuPoolInfo
	modelSize   uint64
	memoryLimit uint64
	preferGPUs  bool
}

// FindBestPlacement finds the best device for a model given current memory state
func FindBestPlacement(ctx *DecisionContext, modelName string) PlacementDecision {
	if ctx == nil || len(ctx.gpuPools) == 0 {
		return PlacementDecision{
			DeviceType:  CPU,
			Reason:      "No GPU pools available or empty",
			CanFitOnGPU: false,
		}
	}

	bestPool := findBestGPUPoolForModel(ctx.gpuPools, ctx.modelSize)

	if bestPool != nil && bestPool.Free >= ctx.modelSize {
		return PlacementDecision{
			DeviceType:  GPU,
			GPUIndex:    bestPool.ID,
			Reason:      "Model fits in GPU memory",
			CanFitOnGPU: true,
		}
	}

	// Model too large for any single GPU - consider CPU or multi-GPU split (future)
	return PlacementDecision{
		DeviceType:  CPU,
		Reason:      "Model exceeds largest GPU memory",
		CanFitOnGPU: false,
	}
}

// findBestGPUPoolForModel finds the best GPU pool for a model based on size and load balancing
func findBestGPUPoolForModel(pools []GpuPoolInfo, modelSize uint64) *GpuPoolInfo {
	if len(pools) == 0 {
		return nil
	}

	bestPool := &pools[0]
	maxUtilization := float64(bestPool.Used) / float64(bestPool.Capacity)

	for _, pool := range pools[1:] {
		utilization := float64(pool.Used) / float64(pool.Capacity)

		// Prefer less utilized pool that can fit the model
		if pool.Free >= modelSize && utilization < maxUtilization {
			maxUtilization = utilization
			bestPool = &pool
		}
	}

	return bestPool
}

// GetOptimalVRAMPerModel calculates optimal VRAM allocation for a model
func GetOptimalVRAMPerModel(modelSize uint64) uint64 {
	// Add ~2GB overhead for KV cache and other runtime allocations
	return modelSize + (2 * 1024 * 1024 * 1024)
}

// FormatDevice returns a friendly device name
func FormatDevice(gpuIndex int, gpuName string) string {
	if gpuName != "" {
		return gpuName
	}
	return "GPU"
}

// GetMemoryUsageMB returns memory usage in MB with 2 decimal precision
func GetMemoryUsageMB(usedBytes uint64) float64 {
	if usedBytes == 0 {
		return 0.0
	}

	bytes := float64(usedBytes)
	mb := bytes / (1024 * 1024)

	// Round to 2 decimal places
	return math.Round(mb*100) / 100
}
