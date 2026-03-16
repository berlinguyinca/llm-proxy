---
# LLM Proxy Administrator Guide
## System Management, Deployment, and Operations Reference

This guide provides comprehensive procedures for administering the LLM Proxy system. It covers deployment, model management, monitoring, scaling, troubleshooting, and operational procedures for DevOps engineers and system administrators.

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Model Deployment Procedures](#model-deployment-procedures)
3. [Scaling Operations](#scaling-operations)
4. [Monitoring & Alerting](#monitoring--alerting)
5. [Troubleshooting Guide](#troubleshooting-guide)
6. [Security Best Practices](#security-best-practices)
7. [Backup & Recovery](#backup--recovery)

---

## Quick Reference

### Common Commands

#### Model Management
```bash
# List all loaded models with details
llm-proxy-manager models list --format table
llm-proxy-manager models list --format json

# Reload a specific model from disk
llm-proxy-manager models reload <model-name>
llm-proxy-manager models reload --all  # Reload all models

# Check if a model is loaded
llm-proxy-manager check <model-name>

# Unload a model gracefully
llm-proxy-manager unload <model-name>
```

#### Routing Operations
```bash
# Inspect routing configuration
llm-proxy-manager routing inspect --format table
llm-proxy-manager routing inspect --format json
```

#### Backend Management
```bash
# Add a new backend for a specific model
llm-proxy-manager backends add <url> --model <name>

# Remove a backend from routing
llm-proxy-manager backends remove <url>
```

#### Health Checks
```bash
# Check overall proxy health
llm-proxy-manager health

# Verify model status
llm-proxy-manager check <model-name>
```

### Opencode Integration
```bash
# Initialize Opencode agent configuration
llm-proxy-manager opencode init --proxy-url http://localhost:9999
llm-proxy-manager opencode init  # Uses default URL

# List Opencode configuration
llm-proxy-manager opencode list
```

### Rate Limiting Configuration

Rate limiting is controlled via environment variables and the `models.yaml` configuration file.

**Environment Variables:**
```bash
RATE_LIMIT_ENABLED=true          # Enable rate limiting (default: false)
RATE_LIMIT_TOKENS=100            # Tokens per request limit
RATE_LIMIT_REFILL_RATE=10        # Tokens per second refill rate
```

**Configuration File (.env.example):**
```bash
# Rate Limiting Settings
RATE_LIMIT_ENABLED=true         # Enable token bucket rate limiting
RATE_LIMIT_TOKENS=100           # Maximum tokens per request (default: 100)
RATE_LIMIT_REFILL_RATE=10       # Tokens per second refill rate (default: 10)
```

**Rate Limit Metrics:**
Monitor rate limit usage via `/metrics` endpoint:
```
rate_limiter_requests_total{name="api"}  # Total requests
rate_limiter_requests_rejected{reason="limit_exceeded"}  # Rejected requests
```

---

## Model Deployment Procedures

### Loading Models Manually

#### Procedure 1: Load a Specific Model
```bash
# Download and load model to proxy memory pool
curl -X POST http://localhost:9999/models/<model-name>/load \
  --data "device=auto"
  
# Example: Load Qwen2.5-7B-Instruct with auto device placement
curl -X POST http://localhost:9999/models/qwen2.5-7b-instruct/load \
  --data 'device=auto'
```

#### Procedure 2: Configure Model Resource Requirements

Edit `config/models.yaml`:
```yaml
# models.yaml - Model configuration with resource hints
models:
  # Standard model entry
  - name: qwen2.5-7b-instruct
    path: ./models/Qwen2.5-7B-Instruct-Q4_K_M.gguf
    min_memory_mb: 6000      # Minimum RAM requirement (when loaded)
    vram_mb_hint: 4096       # Suggested VRAM allocation (if GPU available)
    eviction_priority: 1      # Higher value = evicted first when memory pressure
    discovery_enabled: true   # Allow auto-discovery from LM Studio registry
    
  - name: llama3.2-1b
    path: ./models/Llama-3.2-1B-Instruct-Q4_K_M.gguf
    min_memory_mb: 2000
    eviction_priority: 5      # High priority for eviction (runs on CPU)
    
  - name: mistral-7b-instruct
    path: ./models/Mistral-7B-Instruct-v0.3-Q4_K_M.gguf
    min_memory_mb: 8000
    eviction_priority: 1
```

**Resource Hint Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `min_memory_mb` | int | N/A | Minimum RAM required when model is loaded. Used for auto-load decisions. |
| `vram_mb_hint` | int | N/A | Suggested VRAM allocation. Used for GPU device placement preference. |
| `eviction_priority` | int | 1 | Higher values = model evicted first during memory pressure (1-10 scale). |
| `discovery_enabled` | bool | true | Enable/discover model via `/models/discover` endpoint. |

### Automatic Model Loading (Auto-Load Feature)

The auto-load feature automatically loads models on startup based on resource hints and available memory.

**Enable Auto-Load:**
```bash
# In .env or cmd/proxy/main.go configuration
ENABLE_AUTO_LOAD=true         # Enable automatic model loading on startup
AUTO_LOAD_THRESHOLD_PCT=80    # Load models when pool utilization < 80%
MAX_CONCURRENT_LOADS=2        # Maximum simultaneous model loads
```

**Auto-Load Logic:**
1. On server startup, scan `models.yaml` for models with `min_memory_mb` hints
2. Check available memory pool capacity vs total required memory
3. Load models in order of highest eviction_priority first (most important)
4. Respect GPU device allocation constraints
5. Skip models that would exceed memory thresholds

**Verify Auto-Load:**
```bash
# After startup, check which models are loaded
curl http://localhost:9999/models/stats | jq '.models[] | {name: .name, status: .status}'

# Expected output for auto-loaded models:
[
  {"name": "llama3.2-1b", "status": "loaded"},
  {"name": "mistral-7b-instruct", "status": "loaded"},
  {"name": "gemma-2b-it", "status": "unloaded"}  # Not loaded due to memory constraints
]
```

### Device Placement Configuration

#### CPU vs GPU Placement Decisions

The system automatically decides device placement based on:
1. Available VRAM from NVIDIA GPUs
2. Model size requirements
3. User-specified preferences in config

**Configuration:**
```yaml
# In cmd/proxy/main.go or .env
DEFAULT_DEVICE="cpu"  # Default when no GPU available
GPU_PRIORITY=high     # Prefer GPU placement when possible
```

#### GPU-Specific Configuration

For NVIDIA GPU usage:
```bash
# Check GPU detection
nvidia-smi -L  # List available GPUs
nvidia-smi --query-gpu=index,name,memory.total,memory.free --format=csv

# Output example:
# GPU  Index   Name                       Memory-Total  Memory-Free
#       0      GeForce RTX 4090          24576 MiB     18432 MiB
```

---

## Scaling Operations

### Horizontal Scaling with Multiple Proxy Instances

#### Architecture Setup

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   LB        │────▶│  Proxy-1    │────▶│    LM Studio│
└─────────────┘     └─────────────┘     └─────────────┘
                         │                  │
                         ▼                  ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  Proxy-2    │◀───▶│   Proxy-3   │
                    └─────────────┘     └─────────────┘
```

#### Step-by-Step Scaling Guide

**1. Prepare Additional Proxy Instances:**
```bash
# Clone and build additional proxy instances
cd /opt/llm-proxy
make build

# Or using Docker directly
docker build -t llm-proxy:latest .

# Start additional instance with unique port
docker run -d \
  --name llm-proxy-2 \
  --env PROXY_PORT=9998 \
  --env MEMORY_POOL_SIZE_GB=8 \
  llm-proxy:latest
```

**2. Configure Shared Memory Pool (Optional):**

For multi-instance setups with shared eviction policy:
```yaml
# In each instance's config/models.yaml
models:
  - name: shared-model-1
    path: /opt/models/shared/qwen2.5-1b.gguf
    min_memory_mb: 1000
    eviction_priority: 10
    discovery_enabled: false  # Disable discovery for shared models
```

**3. Load Balance Traffic:**

Using Nginx as reverse proxy:
```nginx
# /etc/nginx/conf.d/llm-proxy.conf
upstream llm_proxy_backend {
    least_conn;
    server http://proxy1:9999;
    server http://proxy2:9998;
}

server {
    listen 80;
    
    location / {
        proxy_pass http://llm_proxy_backend/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        
        # Health check for load balancer
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

#### Load Balancer Health Checks

Configure health checks to use the `/health` endpoint:
```nginx
location /health {
    proxy_pass http://llm_proxy_backend/health;
    access_log off;  # No logs for health checks
}
```

### Memory Pool Configuration

**Calculate Optimal Memory Pool Size:**

```bash
#!/bin/bash
# script: calculate-memory-requirements.sh

MODELS=(
  "qwen2.5-7b:7GB"
  "llama3.2-1b:2GB"
  "mistral-7b:8GB"
)

TOTAL_RAM_MIB=$(($(grep MemTotal /proc/meminfo | awk '{print $2}') / 1024))
AVAILABLE_RAM=$((TOTAL_RAM_MIB - 5120))  # Reserve 5GB for system

echo "Total RAM: ${TOTAL_RAM_MIB}MB"
echo "Available after reservation: ${AVAILABLE_RAM}MB"

# Calculate model requirements
required=0
for model in "${MODELS[@]}"; do
  name=$(echo $model | cut -d: -f1)
  mem=$(echo $model | cut -d: -f2)
  required=$((required + mem))
done

echo "Model memory requirement: ${required}MB"
echo "Remaining after models: $((AVAILABLE_RAM - required))MB"
```

**Configure Memory Pool:**
```bash
# In .env or cmd/proxy/main.go
MEMORY_POOL_SIZE_GB=12    # 12GB total pool (adjust based on calculation)
SWAP_ENABLED=false        # Disable swap for faster responses
```

---

## Monitoring & Alerting

### Prometheus Metrics Endpoint

The `/metrics` endpoint exposes Prometheus-compatible metrics:

#### Request Metrics
```
# Total requests processed
proxy_requests_total{method="POST",endpoint="/v1/chat/completions"} 3421

# Request latency histogram (in milliseconds)
proxy_request_latency_seconds_bucket{le="0.1"} 156
proxy_request_latency_seconds_bucket{le="0.5"} 892
proxy_request_latency_seconds_bucket{le="1.0"} 2103
proxy_request_latency_seconds_bucket{le="+Inf"} 3421

# Rate limiting metrics
rate_limiter_requests_total{name="api"} 1247
rate_limiter_requests_rejected{reason="limit_exceeded"} 23
```

#### Health Status Metrics
```
# Overall health status (1=healthy, 0=unhealthy)
health_check_status 1

# Model memory utilization
memory_pool_utilization 0.735  # 73.5% utilized

# GPU metrics (if available)
gpu_memory_utilization 0.82  # 82% VRAM utilization
evictions_pending 2  # Models waiting to be evicted due to memory pressure
```

### Grafana Dashboard Integration

#### Dashboard Configuration

Place `dashboards/llm-proxy.json` in Grafana provisioning:
```bash
# In docker-compose.yml, add volume mount for dashboards
volumes:
  - ./dashboards:/etc/grafana/provisioning/dashboards
  - ./dashboards-dashboardproviders:/etc/grafana/provisioning/dashboards/dashboardproviders
  - ./alert_rules:/etc/grafana/provisioning/alerting/rules
```

#### Key Queries for Grafana

**Memory Pool Usage:**
```promql
# Memory pool utilization percentage
(max by (instance) (memory_pool_utilization_bytes)) / 
(sum by (instance) (memory_pool_capacity_bytes)) * 100
```

**Request Rate Over Time:**
```promql
rate(proxy_requests_total[5m])
```

**Error Rate:**
```promql
increase(proxy_errors_total[1h]) / 
increase(proxy_requests_total[1h]) * 100
```

### Alert Rules

Place `alert_rules.yml` in monitoring configuration:

```yaml
# alert_rules.yml for Prometheus alertingmanager rules
groups:
  - name: llm-proxy-alerts
    interval: 30s
    rules:
      # High memory pressure
      - alert: HighMemoryPressure
        expr: memory_pool_utilization > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory pressure detected"
          description: "Memory pool utilization is above 85% on {{ $labels.instance }}"

      # Rate limit exhaustion
      - alert: RateLimitExhaustion
        expr: rate_limiter_requests_rejected / rate_limiter_requests_total > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Rate limit exhaustion detected"
          description: "{{ $value }}% of requests being rejected due to rate limits"

      # High request latency
      - alert: HighRequestLatency
        expr: histogram_quantile(0.95, 
          sum(rate(proxy_request_latency_seconds_bucket[5m])) by (le)
        ) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High 95th percentile latency"
          description: "P95 request latency is {{ $value }}s, above 10s threshold"
```

---

## Troubleshooting Guide

### Common Error Messages and Fixes

#### Error: "No models currently loaded"

**Cause:** No models configured in `models.yaml` or models not downloaded.

**Fix:**
```bash
# Add model configuration to config/models.yaml
nano config/models.yaml

# Download required model files
cd ./models
wget https://huggingface.co/Qwen/Qwen2.5-7B-Instruct/resolve/main/Qwen2.5-7B-Instruct-Q4_K_M.gguf -O qwen2.5-7b-instruct.gguf

# Verify model file integrity
ls -lh ./models/qwen2.5-7b-instruct.gguf
```

#### Error: "GPU detection failed"

**Cause:** NVIDIA driver not installed or GPU not detected.

**Fix:**
```bash
# Check if NVIDIA kernel modules are loaded
lsmod | grep nvidia

# Install NVIDIA driver (Ubuntu/Debian):
apt-get install nvidia-driver-535

# Verify GPU detection
nvidia-smi

# Expected output:
# +-----------------------------------------------------------------------------------------+
| GPU Name                   | Memory-Total | Memory-Free |  PCI Bus ID    |
|-----------------------------+--------------+--------------+----------------|
| NVIDIA GeForce RTX 4090     |   24576 MiB  |   18432 MiB  |    00000000:0b|
+-----------------------------------------------------------------------------------------+
```

#### Error: "Model file not found"

**Cause:** Model path doesn't exist or model not downloaded.

**Fix:**
```bash
# Check models directory
ls -la ./models/

# Download specific model from Hugging Face
cd ./models
huggingface-cli download Qwen/Qwen2.5-7B-Instruct Qwen2.5-7B-Instruct-Q4_K_M.gguf --local-dir .

# Verify file exists and has correct size
ls -lh Qwen2.5-7B-Instruct-Q4_K_M.gguf  # Should be ~3-4GB for quantized model
```

#### Error: "Rate limit exceeded" (HTTP 429)

**Cause:** Too many requests or request too large.

**Fix Options:**

1. **Increase rate limit in environment variables:**
   ```bash
   RATE_LIMIT_TOKENS=1000      # Increase from default 100
   RATE_LIMIT_REFILL_RATE=50   # Refill faster
   ```

2. **Disable rate limiting for trusted clients:**
   ```yaml
   # In models.yaml
   - name: internal-api
     min_memory_mb: 500
     rate_limit_disabled: true  # No rate limiting for this model
   ```

3. **Use larger request window via sliding window:**
   ```bash
   # Configure in .env (requires custom implementation)
   RATE_LIMIT_WINDOW_SECONDS=60
   RATE_LIMIT_MAX_REQUESTS_PER_WINDOW=100
   ```

### Memory Pressure Troubleshooting

**Symptoms:** Frequent evictions, slow responses, high latency.

**Diagnosis:**
```bash
# Check memory pool utilization
curl http://localhost:9999/health | jq '.memory_utilization'

# Expected: < 0.85 for healthy operation
{"status":"ok","memory_utilization":0.73}  # Good
{"status":"ok","memory_utilization":0.92}  # Warning - consider adding memory or reducing models
```

**Solutions:**

1. **Evict unused models:**
   ```bash
   for model in $(curl -s http://localhost:9999/models/stats | jq -r '.models[] | select(.status=="loaded") | .name'); do
     # Check if model has been used recently (last 24h)
     last_used=$(curl -s "http://localhost:9999/models/${model}/stats" | jq -r '.last_used')
     if [[ "${last_used}" == "never" ]]; then
       echo "Evicting unused model: ${model}"
       curl -X DELETE "http://localhost:9999/models/${model}"
     fi
   done
   ```

2. **Increase memory pool size:**
   ```bash
   # In .env or main.go
   MEMORY_POOL_SIZE_GB=16    # Increase from 12GB
   SWAP_ENABLED=true         # Enable swap if RAM expansion not possible
   ```

3. **Add more proxy instances for horizontal scaling** (see Scaling Operations section)

### Rate Limiter Debugging

**Enable debug logging:**
```bash
# In .env
LOG_LEVEL=debug

# Or in main.go configuration
logging.level = "debug"
```

**Check rate limit metrics:**
```bash
curl http://localhost:9999/metrics | grep rate_limiter
```

---

## Security Best Practices

### API Key Management

#### Generate Secure API Keys

Opencode generates API keys for agent integration:
```bash
# Initialize configuration (generates new API key)
llm-proxy-manager opencode init --proxy-url http://localhost:9999

# The output includes the generated API key:
# Generated API key: sk-opencode-dev-2026-03-15-a7f3c9e2d8b4...
```

#### Rotate API Keys Regularly

When updating Opencode configuration:
```bash
# Remove old config
rm .opencode/models.yaml

# Generate fresh configuration with new API key
llm-proxy-manager opencode init --proxy-url http://localhost:9999
```

### Environment Variable Protection

**Never commit secrets to version control:**

Create `.env.local` for local development (not committed):
```bash
# .env.local (DO NOT ADD TO GIT)
PROXY_PORT=9999
MEMORY_POOL_SIZE_GB=12
RATE_LIMIT_ENABLED=true
RATE_LIMIT_TOKENS=100
API_KEY=sk-my-secret-api-key-do-not-commit

# Add to .gitignore:
.env.local
.env.development
```

### Network Security

**TLS/SSL Configuration:**

For production deployments, enable HTTPS:
```bash
# Generate self-signed certificate (for testing)
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout key.pem -out cert.pem -days 365

# In cmd/proxy/main.go or .env
TLS_ENABLED=true
TLS_CERT_PATH=./certs/cert.pem
TLS_KEY_PATH=./certs/key.pem
```

**Rate Limiting for API Protection:**

Always enable rate limiting in production:
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_TOKENS=100
RATE_LIMIT_REFILL_RATE=10
```

---

## Backup & Recovery

### Model Configuration Backup

```bash
#!/bin/bash
# script: backup-configuration.sh

BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup models.yaml configurations
cp config/models.yaml "$BACKUP_DIR/models-${TIMESTAMP}.yaml"
cp .opencode/models.yaml "$BACKUP_DIR/.opencode-${TIMESTAMP}.yaml" 2>/dev/null || true

# Backup environment variables (without secrets)
grep -v "API_KEY=" .env > "$BACKUP_DIR/.env-no-secrets-${TIMESTAMP}"

echo "Backup completed: $BACKUP_DIR"
```

### Model Data Recovery

If a model file is corrupted or deleted:

**From Original Source:**
```bash
# Download from Hugging Face directly
cd ./models
huggingface-cli download Qwen/Qwen2.5-7B-Instruct Qwen2.5-7B-Instruct-Q4_K_M.gguf --local-dir .

# Or use wget/curl with direct link
wget "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct/resolve/main/Qwen2.5-7B-Instruct-Q4_K_M.gguf" \
  -O ./Qwen2.5-7B-Instruct-Q4_K_M.gguf
```

**From Backup:**
```bash
# Restore from backup
cp ./backups/models-20260314-103045.yaml ./config/models.yaml
cp ./backups/.opencode-20260314-103045.yaml .opencode/models.yaml

# Restart proxy to reload models
curl -X POST http://localhost:9999/reload
```

### Quick Recovery Checklist

1. ✅ Backup current configuration (`.env`, `models.yaml`)
2. ✅ Check model files are present (`./models/` directory)
3. ✅ Verify proxy is running (`docker ps` or `ps aux | grep llm-proxy`)
4. ✅ Restart proxy: `docker restart llm-proxy` or `go run ./cmd/proxy &`
5. ✅ Health check: `curl http://localhost:9999/health`
6. ✅ Check loaded models: `curl http://localhost:9999/models/stats`

---

## Appendix A: Architecture Overview

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        LLM Proxy System                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐   │
│  │   HTTP API   │▶ │  Router      │▶ │  Model Manager      │   │
│  │ /v1/chat/*   │  │ Path-based   │  │ Memory pool         │   │
│  └──────────────┘  └──────────────┘  └─────────────────────┘   │
│       │                │                  │                      │
│       ▼                ▼                  ▼                      │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Memory Pool                          │   │
│  │  ┌──────────────────────────────────────────────┐      │   │
│  │  │ Loaded Models (with eviction policy)         │      │   │
│  │  ├──────────────┬──────────────┬───────────────┤      │   │
│  │  │ Model-1     │  Model-2     │  Model-N     │      │   │
│  │  │ (GPU/CPU)   │  (GPU/CPU)   │  (GPU/CPU)   │      │   │
│  │  └──────────────┴──────────────┴───────────── ─┘      │   │
│  └─────────────────────────────────────────────────────────┘   │
│       ▲                ▲                  ▲                      │
│       │                │                  │                      │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐   │
│  │ Health Endp. │  │   Metrics    │  │  Rate Limiter      │   │
│  │ /health      │  │ /metrics     │  │ Token bucket       │   │
│  └──────────────┘  └──────────────┘  └─────────────────────┘   │
│                         │                                        │
│              ┌──────────┴──────────┐                             │
│              │   Discovery Endpoint│                             │
│              │  /models/discover   │                             │
│              └─────────────────────┘                             │
└─────────────────────────────────────────────────────────────────┘

External Systems:
├── LM Studio servers (model serving)
├── Opencode agents (local AI configuration)
├── Prometheus/Grafana (monitoring)
└── Load balancers (Nginx, HAProxy, etc.)
```

### Data Flow

```
Client Request → Nginx LB → LLM Proxy API
                              │
                              ▼
                      Router (path-based)
                              │
                              ▼
                        Model Manager
                         │
                ┌────────┴────────┐
                │                 │
          Loaded Models     Need to Load?
             │                  │
         Serve Request    Check Resource Hints
                            │
                      ┌────┴────┐
              Sufficient?│        │
                Yes      │ No    │
              ┌──────┐  └────┬───┘
              │ Serve │     │
              └──────┘     ▼
                     Load Model from Disk
                          │
                   Update Eviction Priority
```

---

## Appendix B: Environment Variable Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `PROXY_PORT` | 9999 | HTTP server port |
| `MEMORY_POOL_SIZE_GB` | N/A (auto) | Total memory pool size in GB |
| `SWAP_ENABLED` | false | Enable swap file for overflow handling |
| `ENABLE_AUTO_LOAD` | true | Auto-load models on startup |
| `AUTO_LOAD_THRESHOLD_PCT` | 80 | Load when pool utilization < this % |
| `MAX_CONCURRENT_LOADS` | 2 | Max simultaneous model loads |
| `RATE_LIMIT_ENABLED` | false | Enable rate limiting |
| `RATE_LIMIT_TOKENS` | 100 | Tokens per request limit |
| `RATE_LIMIT_REFILL_RATE` | 10 | Tokens/second refill rate |
| `TLS_ENABLED` | false | Enable HTTPS/TLS |
| `TLS_CERT_PATH` | N/A | Path to TLS certificate file |
| `TLS_KEY_PATH` | N/A | Path to TLS key file |
| `LOG_LEVEL` | info | Logging level (debug, info, warn, error) |

---

## Appendix C: Quick Commands Reference

### Daily Operations

```bash
# Check system health
curl http://localhost:9999/health | jq .

# List loaded models
llm-proxy-manager models list

# View routing configuration  
llm-proxy-manager routing inspect

# Reload all models from disk (after updating)
llm-proxy-manager models reload --all

# Add new model backend
llm-proxy-manager backends add http://localhost:1234/v1/chat/completions \
  --model qwen2.5-7b-instruct

### Model Management Commands

```bash
# Load a specific model
curl -X POST http://localhost:9999/models/<name>/load

# Unload a model (free up memory)
curl -X DELETE http://localhost:9999/models/<name>

# Check model status
llm-proxy-manager check <model-name>

# List Opencode configuration
llm-proxy-manager opencode list
```

### Emergency Commands

```bash
# Force reload all models (after memory issues)
curl -X POST http://localhost:9999/reload --data 'force=true'

# Evict all loaded models (free full pool)
curl -X DELETE http://localhost:9999/models/stats | \
  jq -r '.models[] | select(.status=="loaded") | "\(.name)"' | \
  xargs -I {} curl -X DELETE "http://localhost:9999/models/{}"

# Health diagnostic
curl http://localhost:9999/health | jq .

# Full metrics dump
curl http://localhost:9999/metrics > /tmp/prometheus_metrics.txt
```

---

## Appendix D: Glossary

| Term | Definition |
|------|------------|
| **Model Pool** | Shared memory region where models are loaded and evicted based on usage |
| **Eviction Priority** | Number (1-10) indicating how urgently a model should be unloaded when memory is low |
| **Rate Limiting** | Token bucket algorithm controlling request frequency to prevent abuse |
| **Discovery Endpoint** | `/models/discover` exposes loaded models for Opencode agents to find them |
| **Auto-Load** | Automatic model loading on startup based on resource hints in config |
| **VRAM Hint** | Suggested GPU memory allocation for optimal performance |

---

*Last updated: March 15, 2026*
*Version: 1.0.0*
