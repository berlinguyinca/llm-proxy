# Management CLI Architecture

## Overview

The LLM Proxy management tool (`cmd/management/main.go`) provides a CLI interface for operational control of the proxy server without requiring service restarts. All operations are implemented inline in the CLI rather than as a separate package, keeping dependencies minimal and self-contained.

## Core Types (Embedded in CLI)

### Manager

Embeds HTTP client for model-level management operations:

- **`NewManager(proxyURL string)`**: Creates manager with HTTP client
- **`ListModels() ([]ModelInfo, error)`**: Fetches `/models/stats` endpoint data
- **`ReloadModel(name string) error`**: POST to `/models/{name}/reload`
- **`UnloadModel(name string) error`**: DELETE `/models/{name}`
- **`ReloadAll() error`**: Reload all models sequentially
- **`GetModelStatus(name string) (*ModelInfo, error)`**: Find model in list
- **`CheckModelStatus(name string) error`**: Verify model is loaded

### BackendManager (Simplified)

Embeds HTTP client for health checking:

- **`GetRoutingMap() ([]RoutingEntry, error)`**: Fetches `/models/stats` endpoint data
- **`AddBackend(modelName, baseURL string)`**: Logs backend addition (future: update routing)
- **`RemoveBackend(url string)`**: Logs backend removal (future: update routing)

## Data Structures (Embedded in CLI)

### ModelInfo

```go
type ModelInfo struct {
    Name    string `json:"name"`
    Path    string `json:"path"`
    Device  string `json:"device"` // "cpu", "gpu_0", etc.
    RAM_MB  int64  `json:"ram_mb"`
    VRAM_MB int64 `json:"vram_mb"`
}
```

### RoutingEntry

```go
type RoutingEntry struct {
    ModelName       string  `json:"model_name"`
    BackendURL      string  `json:"backend_url"`
    DiscoveryEnabled bool   `json:"discovery_enabled"`
    Status          string  `json:"status"` // "healthy", "degraded"
}
```

## Wire Connections

| Manager Operation | Proxy Endpoint | Method | Timeout |
|-------------------|---------------|--------|---------|
| List models | `/models/stats` | GET | 30s |
| Reload model | `/models/{name}/reload` | POST | 60s |
| Unload model | `/models/{name}` | DELETE | 30s |

| BackendManager Operation | Proxy Endpoint | Method | Timeout |
|--------------------------|---------------|--------|---------|
| Get routing map | `/models/stats` | GET | 30s |

## Output Formats

Both managers support two output formats:
- `table`: Human-readable ASCII table (default)
- `json`: Machine-parseable JSON array

Control via:
1. `--format <table\|json>` CLI flag (global flag)
2. `OUTPUT_FORMAT=<format>` environment variable

## Health Check

The health command validates proxy connectivity:
- Endpoint: `/health`
- Method: GET
- Timeout: 5s
- Success criteria: HTTP 200 with body content

## Usage Patterns

### List Loaded Models (Table)
```bash
llm-proxy-manager models list
```

### List Loaded Models (JSON)
```bash
llm-proxy-manager models list --format json
# or
OUTPUT_FORMAT=json llm-proxy-manager models list
```

### Reload Specific Model
```bash
llm-proxy-manager models reload qwen2.5-7b-chat
```

### Unload Model
```bash
llm-proxy-manager models unload qwen2.5-7b-chat
```

### Check Routing Configuration (Table)
```bash
llm-proxy-manager routing show
```

### Check Routing Configuration (JSON)
```bash
llm-proxy-manager routing show --format json
```

### Reload All Models
```bash
llm-proxy-manager models reload --all
```

### Add Backend for LM Studio
```bash
llm-proxy-manager backends add http://localhost:1234/v1/chat/completions --model qwen2.5-7b-chat
# Note: Current implementation logs the action; routing config persisted via proxy /models/stats endpoint
```

### Check Proxy Health
```bash
llm-proxy-manager health
# Output: ✓ Proxy is healthy
```

### Verify Model Status
```bash
llm-proxy-manager check qwen2.5-7b-chat
# Output: ✓ Model qwen2.5-7b-chat is loaded
```

## Implementation Notes

- **Single Binary**: All management functionality contained in one CLI binary (`bin/llm-proxy-manager`)
- **No External Dependencies**: Uses only `github.com/spf13/cobra` for CLI parsing
- **REST API Integration**: Leverages existing proxy endpoints; no separate management server needed
- **Future-Proof Design**: BackendManager has placeholder implementations for future routing config updates

## Future Extensions

- Persist routing config to JSON file with `--config` flag
- Support batch operations (reload/unload multiple models at once)
- Add model discovery via HTTP probes from CLI
- Implement graceful shutdown handlers for unload operations
- Add verbose mode (`-v`) for detailed operation logging
