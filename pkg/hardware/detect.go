// Package hardware provides GPU detection and monitoring capabilities
package hardware

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GpuDevice represents a GPU device with its capabilities
type GpuDevice struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	MemoryTotal    int64  `json:"memory_total_bytes"`
	MemoryFree     int64  `json:"memory_free_bytes"`
	ComputeCapable bool   `json:"compute_capable"`
	IsIntegrated   bool   `json:"is_integrated"`
}

// HardwareInfo contains information about system hardware
type HardwareInfo struct {
	GPUs        []GpuDevice `json:"gpus"`
	CpuCores    int         `json:"cpu_cores"`
	MemoryTotal uint64      `json:"memory_total_bytes"`
	MemoryFree  uint64      `json:"memory_free_bytes"`
}

// DetectGPUs discovers available NVIDIA GPUs
func DetectGPUs() ([]GpuDevice, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,memory.free,memory.used")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil // No NVIDIA GPU means gracefully handle on CPU only
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpus []GpuDevice

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			continue
		}

		name := parts[0]
		memTotalStr := strings.TrimSpace(strings.ReplaceAll(parts[1], "MiB", ""))
		memFreeStr := strings.TrimSpace(strings.ReplaceAll(parts[2], "MiB", ""))

		memTotal, _ := parseMemory(memTotalStr)
		memFree, _ := parseMemory(memFreeStr)

		gpus = append(gpus, GpuDevice{
			Name:           name,
			MemoryTotal:    memTotal,
			MemoryFree:     memFree,
			ComputeCapable: true,
			IsIntegrated:   false,
		})
	}

	return gpus, nil
}

// GetHardwareInfo gets comprehensive hardware information
func GetHardwareInfo() (*HardwareInfo, error) {
	gpus, err := DetectGPUs()
	if err != nil {
		return &HardwareInfo{GPUs: []GpuDevice{}}, nil
	}

	// Get system memory (placeholder - would need OS-specific implementation)
	memTotal := uint64(16 * 1024 * 1024 * 1024) // 16GB as placeholder
	cpuCores := runtime.NumCPU()

	return &HardwareInfo{
		GPUs:        gpus,
		CpuCores:    cpuCores,
		MemoryTotal: memTotal,
		MemoryFree:  memTotal / 4, // Rough estimate - free is typically 25-50% of total
	}, nil
}

// parseMemory converts memory string to bytes
func parseMemory(s string) (int64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, "MiB", ""))
	var val int64
	var unit int

	switch {
	case strings.HasSuffix(s, "MB"):
		val, _ = strconv.ParseInt(strings.TrimSuffix(s, "MB"), 10, 64)
		unit = 2 // MB to bytes
	case strings.HasSuffix(s, "GB"):
		val, _ = strconv.ParseInt(strings.TrimSuffix(s, "GB"), 10, 64)
		unit = 3 // GB to bytes
	default:
		val, _ = strconv.ParseInt(s, 10, 64)
		unit = 0 // Assume bytes if no unit
	}

	if unit == 2 {
		return val * 1024 * 1024, nil
	}
	if unit == 3 {
		return val * 1024 * 1024 * 1024, nil
	}
	return val, nil
}
