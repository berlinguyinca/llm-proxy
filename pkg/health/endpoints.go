// Package health provides health check endpoint implementations
package health

import (
	"encoding/json"
	"fmt"
)

// HealthStatus represents the overall health status of the proxy
type HealthStatus string

const (
	Healthy   HealthStatus = "healthy"
	Degraded  HealthStatus = "degraded"
	Unhealthy HealthStatus = "unhealthy"
)

// ModelStats represents statistics for all loaded models
type ModelStats struct {
	TotalLoaded int `json:"total_loaded"`
}

// ModelInfo contains information about a model
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	QualifiedName string `json:"qualified_name"`
	Device        string `json:"device"`
	MemorySize    uint64 `json:"memory_size_bytes"`
	Status        Status `json:"status"`
}

// Status represents the current status of a model
type Status string

const (
	Loaded   Status = "loaded"
	Unloaded Status = "unloaded"
)

// HealthReport holds comprehensive health check information
type HealthReport struct {
	Status   HealthStatus `json:"status"`
	Messages []string     `json:"messages"`
}

// GpuStats represents GPU statistics for /gpu/stats endpoint
type GpuStats struct {
	TotalGPUs int       `json:"total_gpus"`
	GPUs      []GpuInfo `json:"gpus"`
}

// GpuInfo contains information about a single GPU
type GpuInfo struct {
	Name        string `json:"name"`
	MemoryTotal uint64 `json:"memory_total_bytes"`
	MemoryFree  uint64 `json:"memory_free_bytes"`
}

// HealthCheck performs comprehensive health checks
func HealthCheck(gpuStats *GpuStats, models []*ModelInfo) (*HealthReport, error) {
	report := &HealthReport{
		Status:   Healthy,
		Messages: []string{},
	}

	// Check GPU health
	if gpuStats == nil || len(gpuStats.GPUs) == 0 {
		report.Messages = append(report.Messages, "No GPUs detected (CPU mode)")
	} else if len(gpuStats.GPUs) > 0 {
		report.Messages = append(report.Messages, fmt.Sprintf("%d GPU(s) detected", len(gpuStats.GPUs)))
	}

	// Check models
	if len(models) == 0 {
		report.Status = Degraded
		report.Messages = append(report.Messages, "No models registered")
	} else {
		report.Messages = append(report.Messages, fmt.Sprintf("%d model(s) registered", len(models)))
		for _, model := range models {
			if model.Status == Loaded {
				report.Messages = append(report.Messages, fmt.Sprintf("Loaded: %s (%s)", model.Name, model.Device))
			} else if model.Status == Unloaded {
				report.Messages = append(report.Messages, fmt.Sprintf("Unloaded: %s (%s)", model.Name, model.Device))
			}
		}
	}

	return report, nil
}

// GenerateModelStats generates statistics for all models
func GenerateModelStats(models []*ModelInfo) (*ModelStats, error) {
	return &ModelStats{
		TotalLoaded: len(models),
	}, nil
}

// FormatMemoryBytes formats bytes to a human-readable string
func FormatMemoryBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// HealthStatusResponse is the response format for /health endpoint
type HealthStatusResponse struct {
	Status   string   `json:"status"`
	Messages []string `json:"messages"`
}

// MarshalHealthResponse marshals health report to JSON with status string
func MarshalHealthReport(report *HealthReport) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"status":   string(report.Status),
		"messages": report.Messages,
	})
}
