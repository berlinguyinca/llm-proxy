# LLM Proxy Operational Guide

This guide covers operational procedures for managing the LLM Proxy at runtime without service restarts.

---

## Table of Contents

1. [CLI Management Interface](#cli-management-interface)
2. [Model Management](#model-management)
3. [Routing Inspection](#routing-inspection)
4. [Backend Management](#backend-management)
5. [Health Checks](#health-checks)
6. [Common Operational Procedures](#common-operational-procedures)
7. [GPU Memory Management](#gpu-memory-management)
8. [Monitoring Model Usage](#monitoring-model-usage)

---

## CLI Management Interface

The LLM Proxy comes with a built-in CLI management tool (`llm-proxy-manager`) that enables operational control without requiring service restarts.

### Installation

The CLI is included in the same Docker image as the proxy binary:

```bash
# Build
docker build -t llm-proxy:latest .

# Run with both binaries
docker run --rm -p 9999:9999 llm-proxy:latest ./proxy

# Or use separate container for management CLI (optional)
docker run --rm -it --network host llm-proxy:latest /bin/sh
```

### Available Commands

**Model Management:**

- `llm-proxy-manager models list` - List all loaded models with device placement (CPU/GPU) and memory usage
- `llm-proxy-manager models reload <name>` - Reload a specific model from disk
- `llm-proxy-manager models unload <name>` - Gracefully unload a model from memory pool
- `llm-proxy-manager models reload --all` - Reload all loaded models

**Routing Inspection:**

- `llm-proxy-manager routing show` - Show which models are served from which proxy backends (LM Studio, Ollama, etc.)

**Backend Management:**

- `llm-proxy-manager backends add <url> --model <name>` - Add a new proxy backend for model routing
- `llm-proxy-manager backends remove <url>` - Remove a configured proxy backend from routing

**Health Checks:**

- `llm-proxy-manager health` - Check overall proxy health
- `llm-proxy-manager check <model>` - Verify that a specific model is currently loaded

### Output Formats

Both table and JSON output formats are supported:

```bash
# Table output (default, human-readable)
$ llm-proxy-manager models list
NAME                      DEVICE   RAM (MB)  VRAM (MB)
qwen2.5-7b-chat          cpu      0         13421

# JSON output (machine-parseable)
$ llm-proxy-manager models list --format json
[
  {
    "name": "qwen2.5-7b-chat",
    "device": "cpu",
    "ram_mb": 0,
    "vram_mb": 13421
  }
]
```

### Common Operational Procedures

**Add a new model dynamically:**

1. Edit `models.yaml` to add the new model path, or
2. Start proxy with `DISCOVERY_ENABLED=true` near LM Studio (model auto-discovered)

3. Reload: `./bin/proxy reload`
4. Verify with CLI: `llm-proxy-manager models list`

**Manage loaded models without restart:**

1. List current models: `llm-proxy-manager models list`
2. Unload unwanted model: `llm-proxy-manager models unload <name>`
3. Reload to pick up changes: `llm-proxy-manager models reload <name>` or `--all`
4. Check model status: `llm-proxy-manager check <name>`

**GPU Memory Management:**

Monitor GPU memory with the stats endpoint:
```bash
# Check GPU usage (in MB)
curl http://localhost:9999/gpu/stats | jq '.total_used_mb, .used_memory_mb'
```

Verify total GPU usage is within threshold: `MEM_THRESHOLD=17000`

**Monitor Model Usage:**

```bash
# List models with memory consumption
$ llm-proxy-manager models list
NAME                      DEVICE   RAM (MB)  VRAM (MB)
qwen2.5-7b-chat          gpu_0    0         13421

# Check model status
$ llm-proxy-manager check qwen2.5-7b-chat
✓ Model qwen2.5-7b-chat is loaded
```

### Cron Job Example (Linux/macOS)

Add periodic model listing to your monitoring system:

```cron
# Add models check every hour
0 * * * * /path/to/bin/llm-proxy-manager models list --format json >> /var/log/proxy-models.log 2>&1
```

### Docker Sidecar Container Example

Run management CLI as a separate container for easier access:

```yaml
services:
  proxy-management:
    build:
      context: .
      dockerfile: Dockerfile
    depends_on:
      - proxy
    environment:
      - PROXY_URL=http://proxy:9999
    command: ["./proxy"]
    volumes:
      - ./models:/data/models  # Shared model storage with main proxy
```

---

## Model Management

### Loading Models

The proxy supports two methods for loading models:

1. **Configuration File** (`models.yaml`): Static configuration
2. **Discovery**: Auto-discover from LM Studio when `DISCOVERY_ENABLED=true`

Example `models.yaml`:
```yaml
- name: qwen2.5-7b-chat
  path: /data/models/qwen2.5-7b-chat-q4_K_M.gguf
```

### Unloading Models

Unload a model to free memory for other models:

```bash
./bin/proxy unload qwen2.5-7b-chat
# or via CLI
llm-proxy-manager models unload qwen2.5-7b-chat
```

### Memory Threshold Enforcement

The proxy enforces combined RAM+VRAM threshold by default (16GB). When exceeded, it evicts least recently used models first.

Adjust threshold in environment:
```bash
MEM_THRESHOLD_GB=20 ./bin/proxy
```

---

## Routing Inspection

View which models are served from which backend servers:

```bash
llm-proxy-manager routing show --format json
```

Output shows model-to-backend mapping with discovery status.

---

## Backend Management

### Adding LM Studio Backends

For direct proxy-to-LM Studio routing without discovery, add a backend:

```bash
llm-proxy-manager backends add http://localhost:1234/v1/chat/completions --model qwen2.5-7b-chat
```

This logs the action; for automatic discovery use `DISCOVERY_ENABLED=true`.

### Removing Backends

Remove a backend from routing configuration:

```bash
llm-proxy-manager backends remove http://localhost:1234/v1/chat/completions
```

Note: Models served by removed backends need to be reloaded or discovered via auto-discovery.

---

## Health Checks

### Overall Proxy Health

```bash
llm-proxy-manager health
# Output: ✓ Proxy is healthy
```

Or check REST API endpoint:
```bash
curl http://localhost:9999/health
```

### Verify Model Status

```bash
llm-proxy-manager check qwen2.5-7b-chat
# Output: ✓ Model qwen2.5-7b-chat is loaded
```

---

## GPU Memory Management

### Check GPU Usage

```bash
curl http://localhost:9999/gpu/stats | jq .
```

Example output:
```json
{
  "total_vram_mb": 15423,
  "used_memory_mb": 10240,
  "devices": [
    {
      "name": "NVIDIA GeForce RTX 4090",
      "device_id": 0,
      "total_vram_mb": 7812,
      "used_vram_mb": 5120
    }
  ]
}
```

### Monitor Threshold

Set `MEM_THRESHOLD_GB` environment variable to control memory allocation. Default is 16GB.

---

## Monitoring Model Usage

The `/models/stats` endpoint provides real-time model information:

```bash
curl http://localhost:9999/models/stats | jq .
```

Or use CLI:
```bash
llm-proxy-manager models list --format json
```

Fields include:
- `name`: Model identifier
- `path`: Filesystem path to model file
- `device`: Device placement (cpu, gpu_0, etc.)
- `ram_mb`: RAM usage in megabytes
- `vram_mb`: GPU VRAM usage in megabytes

---

## Quick Reference

| Task | Command |
|------|---------|
| List loaded models | `llm-proxy-manager models list` |
| Reload model | `llm-proxy-manager models reload <name>` |
| Unload model | `llm-proxy-manager models unload <name>` |
| Check model status | `llm-proxy-manager check <model>` |
| View routing config | `llm-proxy-manager routing show` |
| Check proxy health | `llm-proxy-manager health` |

For JSON output, add `--format json` flag or set `OUTPUT_FORMAT=json`.
