# LLM Proxy Production Deployment Guide

## Quick Start

### Step 1: Set Up Environment Variables

```bash
cp .env.example .env
nano .env  # Edit to set your desired configuration
```

**Key Configuration Options:**
- `PORT=9999` - Proxy listening port
- `RATE_LIMIT_MAX_TOKENS=100` - Maximum global tokens per second
- `DISCOVERY_ENABLED=false` - Enable auto-discovery from LM Studio
- `LOG_LEVEL=info` - Log verbosity
- `METRICS_ENABLED=true` - Expose Prometheus metrics

### Step 2: Build and Run

```bash
# Build the binary
go build -o bin/proxy ./cmd/proxy

# Run with environment file
./bin/proxy --env-file .env
```

**Or use Docker:**
```bash
docker-compose up -d
```

### Step 3: Verify Deployment

```bash
# Health check
curl http://localhost:9999/health

# View model stats
curl http://localhost:9999/models/stats

# Prometheus metrics
curl http://localhost:9999/metrics
```

## Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `ENV_PREFIX` | `` | Prefix for all environment variables |
| `OPENAI_API_KEY` | (empty) | OpenAI API key for proxying requests |
| `ANTHROPIC_API_KEY` | (empty) | Anthropic API key for proxying requests |
| `LM_STUDIO_DISCOVERY_URL` | `http://localhost:1234/api/v1/models` | LM Studio discovery endpoint |
| `RATE_LIMIT_MAX_TOKENS` | `100` | Maximum global tokens per second |
| `RATE_LIMIT_REFILL_RATE` | `10` | Tokens added per second (refill rate) |
| `PORT` | `9999` | Proxy listening port |
| `MEMORY_THRESHOLD_GB` | `16` | Memory pool threshold in GB |
| `DISCOVERY_ENABLED` | `false` | Enable auto-discovery from LM Studio |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `METRICS_ENABLED` | `true` | Expose Prometheus metrics at /metrics |
| `MONITORING_PORT` | `9999` | Port for metrics endpoint |

## Rate Limiting Configuration

Rate limiting protects against abuse by enforcing token budgets:

- **Token Bucket Algorithm**: Tokens refill continuously at configurable rate
- **Global Limits**: Default 100 tokens/second
- **Per-Model Limits**: Can be configured per model in `models.yaml`
- **Response Code**: Returns HTTP 429 when limits exceeded

**Adjust for high-throughput:**
```bash
RATE_LIMIT_MAX_TOKENS=1000      # 1000 tokens/second
RATE_LIMIT_REFILL_RATE=50       # Refill at 50 tokens/second
```

## Memory Management

The proxy manages memory (RAM + GPU VRAM) for loaded models:

- **Threshold**: `MEMORY_THRESHOLD_GB` (default: 16GB combined pool)
- **Eviction Strategy**: LRU (Least Recently Used) when threshold hit
- **Monitoring**: Check `/models/stats` and `/gpu/stats` endpoints

**Increase for heavy workloads:**
```bash
MEMORY_THRESHOLD_GB=32
```

## Docker Compose Deployment

### Basic Setup (Proxy Only)

```yaml
version: '3.8'

services:
  proxy:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: llm-proxy
    ports:
      - "9999:9999"
    environment:
      - PORT=9999
      - RATE_LIMIT_MAX_TOKENS=100
      - DISCOVERY_ENABLED=false
      - LOG_LEVEL=info
    volumes:
      - ./config:/app/config:ro
      - ./logs:/app/logs
    networks:
      - llm-proxy-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9999/health"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  llm-proxy-network:
```

### With LM Studio (Optional)

```yaml
services:
  proxy:
    # ... existing configuration ...

  lm-studio:
    image: lmstudio/lmstudio:latest
    container_name: lm-studio
    ports:
      - "1234:1234"
    volumes:
      - ./lm-studio-data:/data
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]

networks:
  llm-proxy-network:
    external: false
```

Then add models to `config/models.yaml` and enable discovery.

### With Prometheus & Grafana Monitoring

See the included `grafana-docker-compose.yml` file for full monitoring stack setup.

```bash
# Deploy monitoring stack
docker-compose -f grafana-docker-compose.yml up -d

# Access:
# - Proxy API: http://localhost:9999
# - Prometheus UI: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin123)
```

## Prometheus Metrics

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `proxy_requests_total` | Counter | Total requests by status code and model |
| `proxy_request_duration_seconds` | Histogram | Request latency in seconds |
| `proxy_active_connections` | Gauge | Current active connections |
| `registry_model_count` | Gauge | Number of registered models |
| `registry_model_size_bytes_sum` | Counter | Total model size in bytes |

### Access Metrics Endpoint

```bash
curl http://localhost:9999/metrics
```

Or if you set up Grafana monitoring stack, access at:
- Prometheus UI: `http://localhost:9090`
- Grafana: `http://localhost:3000`
- Dashboards: Overview, Performance, Rate Limiting, Resources, Model Stats

## Dockerfile Reference

The multi-stage Dockerfile produces a ~30MB production image:

```dockerfile
FROM golang:1.22-alpine AS builder
# Build stage...

FROM ghcr.io/distribution/static-debian11:latest
# Production runtime...
```

**Benefits:**
- **Size**: ~30MB vs ~900MB for full Go SDK (85% reduction)
- **Security**: Non-root user, distroless base
- **Health Check**: Built into image

## Models Configuration

Add models to `config/models.yaml`:

```yaml
models:
  - name: qwen2.5:7b
    url: http://localhost:1234/v1/chat/completions
    min_memory_mb: 8000
    eviction_priority: 50
    params:
      temperature: 0.7
      top_p: 0.9
```

**Field Descriptions:**
- `min_memory_mb`: Optional - Minimum memory reservation for this model
- `eviction_priority`: Priority level (higher = evicted first when threshold hit)

## Health Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Basic health check |
| `/models/stats` | GET | Model registry with load/unload status |
| `/gpu/stats` | GET | GPU memory and utilization stats |
| `/metrics` | GET | Prometheus metrics |

## Troubleshooting

### High Latency
1. Check model is loaded: `curl http://localhost:9999/models/stats`
2. Verify GPU has VRAM: `nvidia-smi`
3. Review logs for OOM errors
4. Adjust `MEMORY_THRESHOLD_GB` if needed

### Rate Limiting Too Aggressive
Increase token bucket capacity:
```bash
RATE_LIMIT_MAX_TOKENS=500
RATE_LIMIT_REFILL_RATE=25
```

### Models Not Loading from Discovery
1. Verify LM Studio is running: `docker ps | grep lm-studio`
2. Check discovery URL in `.env`: `LM_STUDIO_DISCOVERY_URL`
3. Enable discovery: `DISCOVERY_ENABLED=true`
4. View logs: `docker logs llm-proxy`

### Memory Pool Issues
1. Check current usage: `curl http://localhost:9999/models/stats`
2. Adjust threshold: `MEMORY_THRESHOLD_GB=32`
3. Unload unused models via API or reload proxy
4. Review GPU stats: `docker exec llm-proxy nvidia-smi`

## Security Considerations

### Production Recommendations

1. **TLS/HTTPS**: Configure with reverse proxy (nginx, traefik)
2. **Auth**: Add authentication layer before metrics endpoints
3. **Network Isolation**: Put in private Docker network
4. **Secrets Management**: Use environment secrets or vault for API keys
5. **Resource Limits**: Set memory/CPU limits in docker-compose

Example nginx reverse proxy:
```nginx
server {
    listen 80;
    
    location /metrics {
        proxy_pass http://proxy:9999;
        # Restrict access to internal networks only
        allow 10.0.0.0/8;
        deny all;
    }
    
    location / {
        proxy_pass http://proxy:9999;
        proxy_set_header X-Forwarded-For $remote_addr;
    }
}
```

## Monitoring & Observability

### Grafana Dashboards

Production-ready dashboards are included for key metrics:

- **Overview**: Request rate, throughput, connections
- **Performance**: Latency p50/p95/p99 by model
- **Rate Limiting**: 429 response tracking and token bucket status
- **Resources**: CPU, memory, connection utilization
- **Model Stats**: Model distribution and performance

**Deployment Options:**
1. Docker Compose: Use `grafana-docker-compose.yml`
2. Manual Import: Upload JSON files via Grafana UI
3. Automated Provisioning: Use `provisioning/` directory configs

### Alerting Rules

Alert rules defined in `prometheus-alerts.yml`:
- High error rate (>10%)
- Excessive rate limiting
- High latency (p95 > 5 seconds)
- Memory pressure (>1.5GB)
- Connection overload (>800 connections)

## Scaling Considerations

### Horizontal Scaling
- Deploy multiple proxy instances behind load balancer
- Share model registry state externally if needed
- Each instance has its own memory pool

### Vertical Scaling
- Increase `MEMORY_THRESHOLD_GB` for more models
- Use larger GPU instances for VRAM-intensive workloads
- Monitor `/gpu/stats` for available capacity

## Next Steps & Enhancements

1. **Service Mesh**: Add Istio/Linkerd for advanced observability
2. **Kubernetes**: Convert docker-compose to K8s manifests
3. **Alerting Integration**: Connect alerts to PagerDuty/OpsGenie
4. **Tracing**: Add OpenTelemetry distributed tracing
5. **Logs Aggregation**: Ship logs to Loki/ELK stack

## License & Attribution

Built with Go, Prometheus client library, and Grafana dashboards.

See individual LICENSE files for component licensing details.

## GPU Support & LM Studio Integration

### GPU Configuration in Docker Compose

The `docker-compose.yml` file includes built-in GPU support for both the LLM Proxy and LM Studio services.

#### Requirements for GPU Access:

1. **NVIDIA Driver:**
   ```bash
   # Check driver version
   nvidia-smi
   
   # Expected output should show your GPU (e.g., GeForce RTX, NVIDIA A-series)
   ```

2. **Docker with NVIDIA Runtime:**
   ```bash
   # Install NVIDIA container toolkit
   # Ubuntu/Debian:
   curl -fsSL https://nvidia.github.io/libnvidia-container/libnvidia_container.gpg | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit.gpg \
     && chmod 644 /usr/share/keyrings/nvidia-container-toolkit.gpg \
     && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/nvidia-container-toolkit.gpg] https://nvidia.github.io/libnvidia-container/stable/deb/$(dpkg --print-architecture) /" | sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list \
     && sudo apt-get update && sudo apt-get install -y nvidia-container-toolkit

   # Or use Docker Desktop on Mac/Windows with GPU support enabled
   ```

3. **Enable GPUs for LM Studio (Optional):**
   
   In `docker-compose.yml`, the LM Studio service already has GPU configuration:
   ```yaml
   deploy:
     resources:
       reservations:
         devices:
           - driver: nvidia
             count: 1
             capabilities: [gpu]
   ```

   To enable GPU access for LM Studio, restart the container after installing NVIDIA toolkit.

### Running with and without LM Studio

#### Option 1: Proxy Only (No GPU needed)
```bash
# Just the proxy service
docker-compose up -d proxy
```

#### Option 2: Proxy + LM Studio with GPU
```bash
# Enable LM Studio by uncommenting the lm-studio service in docker-compose.yml
# Then restart:
docker-compose down
docker-compose up -d proxy lm-studio
```

#### Option 3: Full Stack (Proxy + LM Studio + Monitoring)
```bash
# With monitoring stack included
docker-compose up -d
```

### Checking GPU Usage

Once running, check if the proxy is using GPU resources:
```bash
# Check container resource usage
docker stats llm-proxy-proxy

# Should show GPU-MEMORY if proxy is loading models to GPU
```

For LM Studio:
```bash
# Check LM Studio GPU memory usage
nvidia-smi
```

### Troubleshooting GPU Issues

1. **"No NVIDIA devices found" error:**
   - Ensure NVIDIA drivers are installed: `nvidia-smi`
   - Verify Docker has NVIDIA runtime enabled
  
2. **LM Studio won't start with GPU:**
   ```bash
   # Remove container and recreate
   docker-compose down lm-studio
   docker-compose up -d lm-studio
   
   # LM Studio may need to re-detect GPUs after restart
   ```

3. **Models loading to CPU instead of GPU:**
   - Check model size vs available VRAM in `/models/stats` endpoint
   - Ensure `MEMORY_THRESHOLD_GB` allows enough combined RAM + VRAM

### GPU Memory Limits

The proxy respects GPU memory constraints automatically:
- Models larger than available GPU VRAM will load to RAM
- Combined pool (RAM + GPU VRAM) managed by `MEMORY_THRESHOLD_GB`
- LRU eviction removes models from fastest/slowest devices first when needed

See `/models/stats` endpoint for device placement of each loaded model.

---

## GPU Memory Management Best Practices

### Understanding GPU vs CPU Model Loading

The LLM Proxy intelligently manages where models load based on memory availability:

**GPU VRAM Loading:**
- Models fitting entirely in single GPU's VRAM → prefer GPU loading (faster inference)
- Requires `min_memory_mb` in model config or automatic detection
- Best for latency-sensitive applications

**CPU RAM Loading:**
- Models larger than available GPU VRAM → load to system RAM
- Still functional but slower inference (~5-10x slower than GPU)
- Useful for experimentation or testing with large models

### GPU-Specific Tips

1. **Check Available VRAM Before Loading Large Models:**
   ```bash
   nvidia-smi --query-compute-apps=pid,used_memory --format=csv
   # Shows current GPU memory usage per process
   
   nvidia-smi --query-gpu=index,memory.total,memory.used --format=csv
   # Shows total and used VRAM per GPU
   ```

2. **Monitor Model Memory Usage:**
   ```bash
   curl http://localhost:9999/models/stats | jq '.models[] | {name, device, ram_mb, vram_mb}'
   ```

3. **Unload Unused Models to Free GPU VRAM:**
   ```bash
   # Via API endpoint (example for model qwen2.5)
   curl -X DELETE http://localhost:9999/model/qwen2.5
   
   # Or via models.yaml (add to unload config)
   ```

4. **Increase GPU Memory Allocation:**
   - The proxy automatically distributes combined RAM+VRAM pool
   - Set higher `MEMORY_THRESHOLD_GB` for more model loading capacity
   
### Multi-GPU Considerations

The current implementation supports single GPU per container. For multi-GPU setups:

1. **Docker Compose with Multiple GPUs:**
   ```yaml
   deploy:
     resources:
       reservations:
         devices:
           - driver: nvidia
             count: 2  # Use multiple GPUs
             capabilities: [gpu]
   ```

2. **GPU Memory Distribution:**
   - Proxy loads models based on total available VRAM across all GPUs
   - Models may span multiple GPUs automatically if configured

3. **Considerations:**
   - Multi-GPU inference requires specific model configuration (e.g., tensor parallelism)
   - LM Studio supports multi-GPU through its built-in configuration
   - Proxy forwards requests to LM Studio or directly routes to GPU endpoints

### CPU-Only Mode

If you don't have GPU support:

```bash
# Run proxy without Docker GPU features
./bin/proxy --no-gpu

# Or use docker with nvidia runtime removed
docker run --rm llm-proxy:latest \
  -e DISCOVERY_ENABLED=false \
  -e METRICS_ENABLED=true
```

The proxy will automatically detect lack of GPU and load all models to system RAM.

---

## Security Hardening for Production

### Container Isolation

Run containers with limited capabilities:

```bash
docker run --read-only --tmpfs /tmp --cap-drop ALL \
  llm-proxy:latest --no-gpu --metrics-enabled=false
```

### Network Segregation

Use private networks for internal services:
- Proxy on port 9999 (public API)
- Metrics at `localhost:9999/metrics` (internal only)
- LM Studio discovery URL restricted to internal network

### Secrets Management

Avoid hardcoding credentials in `.env`:
- Use Docker secrets or external secret management
- Rotate keys regularly using environment injection
- Never commit secrets to version control

See the full security checklist in our production deployment guide for complete best practices.
