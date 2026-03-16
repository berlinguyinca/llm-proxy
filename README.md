# LLM Proxy

A sophisticated proxy server for managing, loading, and routing Large Language Models (LLMs) with hardware-aware memory management and agent discovery capabilities.

---

## 👋 Quick Start

### 🚀 One-Line Install (Recommended for Quick Start)

Install LLM Proxy, LM Studio, and CLI tools with optional service setup:

```bash
curl -fsSL https://github.com/berlinguyinca/llm-proxy/raw/main/install.sh | bash
```

This script will automatically:
- ✅ Check for and install Go (if needed)
- ✅ Download and configure LM Studio
- ✅ Build the LLM Proxy server binary  
- ✅ Build CLI management tools (`llm-proxy-manager`)
- ✅ Setup Opencode integration
- ⚡ Ask if you want to run as a background service

After installation:
```bash
./bin/proxy              # Start proxy server
curl http://localhost:9999/health  # Verify it's running
llm-proxy-manager models list    # List loaded models
```

### For Developers - Manual Build (5-Minute Setup)

```bash
# 1. Clone repository  
git clone https://github.com/berlinguyinca/llm-proxy.git
cd llm-proxy

# 2. Build the proxy server and CLI tools
go build -o bin/proxy ./cmd/proxy
go build -o bin/llm-proxy-manager ./cmd/management

# 3. Run with default settings (works out of the box!)
./bin/proxy

# 4. Check health
curl http://localhost:9999/health
```

Output should show:
```json
{
  "status": "healthy",
  "total_loaded": 0,
  ...
}
```

Output should show:
```json
{
  "status": "healthy",
  "total_loaded": 0,
  ...
}
```

### For Production - Docker Deployment

```bash
# 1. Copy environment examples
cp .env.example .env

# 2. Edit .env with your settings (see .env.example for all options)

# 3. Build and run
docker-compose up -d

# 4. Check status
docker-compose ps

# 5. View logs
docker-compose logs -f proxy
```

### Essential CLI Commands

```bash
# List loaded models
./llm-proxy-manager models list

# Check health
./llm-proxy-manager health

# Setup Opencode for agents (one-time only)
./llm-proxy-manager opencode init --proxy-url http://localhost:9999

# View all commands
./llm-proxy-manager --help
```

### 🚀 One-Line Install (Recommended for Quick Start)

Install LLM Proxy, LM Studio, and CLI tools with optional service setup:

```bash
curl -fsSL https://github.com/berlinguyinca/llm-proxy/raw/main/install.sh | bash
```

This script will:
- ✅ Check for and install Go (if needed)
- ✅ Download and configure LM Studio
- ✅ Build the LLM Proxy server binary
- ✅ Build CLI management tools  
- ✅ Setup Opencode integration
- ⚡ Ask if you want to run as a background service

After installation:
```bash
./llm-proxy-server  # or start via service
curl http://localhost:9999/health  # Verify it's running
```

## 📦 Installation Options

**Choose your preferred installation method:**

### 🚀 Option 1: One-Line Install (Recommended)
Best for: Getting started quickly with everything included.

```bash
curl -fsSL https://github.com/berlinguyinca/llm-proxy/raw/main/install.sh | bash
```

Includes: LLM Proxy server, CLI tools, LM Studio setup, service configuration.

---

### 🛠️ Option 2: Manual Build (For Developers)  
Best for: Custom configurations, CI/CD pipelines, or when you control Go environment.

See the [Manual Build](#for-developers---manual-build---5-minute-setup) section above.

---

### 🐳 Option 3: Docker Deployment (Production)
Best for: Production environments, consistent deployments, isolated environments.

See the [Docker Deployment](#for-production---docker-deployment) section below.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start) 
  - [One-Line Install](#-one-line-install-recommended-for-quick-start) ⚡ **Recommended**
  - [For Developers - Manual Build](#for-developers---manual-build---5-minute-setup) 🛠️
  - [For Production - Docker Deployment](#for-production---docker-deployment) 🐳
- [CLI Commands](#cli-commands)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Opencode Agent Integration](#opencode-agent-integration)
- [Architecture](#architecture)
- [Development](#development)
- [Troubleshooting](#troubleshooting)


## Features

### 🚀 Core Capabilities

- **Model Management**: Load, unload, and manage multiple LLM models with hardware-aware memory pooling
- **Hardware-Aware Allocation**: Automatic GPU/CPU device placement based on available hardware
- **Rate Limiting**: Token bucket rate limiting per model for fair resource distribution
- **Proxy Routing**: Path-based routing to LM Studio, Ollama, or other LLM backends
- **Model Discovery**: Built-in discovery endpoint for agent integration and monitoring

### 🤖 Agent Integration

- **Opencode Support**: Full Opencode agent discovery and authentication
- **CLI Configuration Management**: Easy setup for agent connections
- **Model Registry**: RESTful API for model information and capabilities

---

## Quick Start

### Prerequisites

- Go 1.21+ 
- LM Studio or compatible LLM server running on `http://localhost:1234`
- NVIDIA GPU (optional, works on CPU-only)

### Installation

```bash
# Clone repository
git clone https://github.com/your-org/llm-proxy.git
cd llm-proxy

# Build the proxy server
go build -o llm-proxy-server ./cmd/proxy

# Build CLI tools
go build -o llm-proxy-manager ./cmd/management
```

### Basic Usage

#### Start the Proxy Server

```bash
# Simple start (default port 9999)
./llm-proxy-server

# Custom port
export PORT=8080
./llm-proxy-server

# Enable model discovery from LM Studio
export DISCOVERY_ENABLED=true
export LM_STUDIO_DISCOVERY_URL=http://localhost:1234/api/v1/models
./llm-proxy-server

# Disable rate limiting for maximum throughput
export DISABLE_RATE_LIMITING=true
./llm-proxy-server
```

#### Load a Model via CLI

```bash
# Register and load a model
./llm-proxy-manager models load \
  --name qwen2.5-7b-chat \
  --path /path/to/model.gguf \
  --device cpu
```

#### View Loaded Models

```bash
./llm-proxy-manager models list
```

#### Check Proxy Health

```bash
curl https://localhost:9999/health
```

---

## 🐳 Docker Deployment (Production-Ready)

### Quick Deploy

```bash
# Copy environment examples
cp .env.example .env

# Edit .env with your settings (optional)
nano .env  # or use your preferred editor

# Start all services (proxy + monitoring stack)
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f proxy

# Stop services
docker-compose down

# Clean up volumes and restart
docker-compose down --volumes && docker-compose up -d
```

### Docker Deployment Options

**Option 1: Development (CPU-only)**
- Best for testing and development
- No GPU passthrough needed
- Models load to CPU memory pool

**Option 2: Production with GPU**
- Requires NVIDIA runtime on Docker host
- Edit docker-compose.yml line 74 to enable GPU devices
- Set `memory` limits higher (e.g., 16GB)

**Option 3: Minimal (Proxy Only)**
```bash
# Build only proxy, run standalone
docker build -t llm-proxy -f Dockerfile .
docker run -p 9999:9999 llm-proxy
```

### Docker Monitoring Stack

When using `docker-compose up -d`, includes:
- **Prometheus** (port 9090) - Metrics collection
- **Grafana** (port 3000) - Visualization dashboards
- Access Grafana at http://localhost:3000 (admin/admin123)

### Docker Health Checks

```bash
# Check proxy health
curl http://localhost:9999/health

# View Prometheus metrics
curl http://localhost:9090/api/v1/query?query=up
```


---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     LLM Proxy Server                          │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                    Memory Pool Manager                    │ │
│  │  - Hardware-aware memory allocation                      │ │
│  │  - GPU/CPU device placement                               │ │
│  │  - Model lifecycle management                             │ │
│  └─────────────────────────────────────────────────────────┘ │
│                              │                                 │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                 Model Registry                            │ │
│  │  - Model registration & tracking                         │ │
│  │  - Load/unload operations                                 │ │
│  │  - Status management                                      │ │
│  └─────────────────────────────────────────────────────────┘ │
│                              │                                 │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                    Proxy Router                           │ │
│  │  - Path-based routing to backends                         │ │
│  │  - Request normalization                                  │ │
│  └─────────────────────────────────────────────────────────┘ │
│                              │                                 │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Rate Limiter (Token Bucket)                  │ │
│  │  - Per-model rate limiting                                │ │
│  │  - Configurable tokens & refill rates                     │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

Exposed Endpoints:
├── GET /health                 - Health check with model stats
├── GET /models/stats           - Model registry information
├── GET /gpu/stats              - GPU hardware information
├── GET /models/discover        - Opencode agent discovery
├── POST /models/load           - Auto-load model from disk
└── /model-*/*                  - Proxy to backend services
```

### Memory Management Strategy

```
Total Memory Pool: threshold (e.g., 16GB)
├── CPU Models      (if no GPU or overflow)
├── GPU Model 0     (primary GPU if available)
└── GPU Model 1     (secondary GPU if available)

Allocation Rules:
- Each model reserves its required VRAM/RAM from pool
- Loading a model checks pool availability first
- Unloading frees memory for other models
- Models can be swapped between CPU/GPU based on hardware
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9999` | Server listen port |
| `MEMORY_THRESHOLD_GB` | `16` | Total memory pool in GB |
| `RATE_LIMIT_MAX_TOKENS` | `100` | Max tokens per request window (when enabled) |
| `RATE_LIMIT_REFILL_RATE` | `10` | Tokens per second refill rate (when enabled) |
| `DISABLE_RATE_LIMITING` | `false` | Set to `true` to completely disable rate limiting |
| `DISCOVERY_ENABLED` | `true` | Enable LM Studio discovery |
| `LM_STUDIO_DISCOVERY_URL` | `http://localhost:1234/api/v1/models` | Discovery endpoint URL |

**To disable rate limiting completely:**
```bash
export DISABLE_RATE_LIMITING=true
./llm-proxy-server
```

This removes all rate limiting overhead and allows unlimited request throughput.

### Model Configuration (Optional)

Create `config/models.yaml`:

```yaml
models:
  - id: qwen-7b
    name: Qwen/Qwen-7B-GGUF/qwen-7b.gguf
    url: http://localhost:1234/v1/chat/completions
    size_gb: 7.0
    device: cpu
    qualified_name: Qwen/Qwen-7B-GGUF/qwen-7b.gguf

  - id: mistral-7b
    name: Mistral/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct.Q4_K_M.gguf
    url: http://localhost:1234/v1/chat/completions
    size_gb: 6.5
    device: cpu
    qualified_name: Mistral/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct.Q4_K_M.gguf
```

Run proxy to auto-load from config:

```bash
export CONFIG_FILE=config/models.yaml
./llm-proxy-server
```

---

## API Reference

### Health Check

**Endpoint:** `GET /health`

Returns server health status and loaded model information.

**Response Example:**
```json
{
  "status": "healthy",
  "total_loaded": 2,
  "models": [
    {
      "id": "qwen-7b",
      "name": "Qwen/Qwen-7B-GGUF/qwen-7b.gguf",
      "device": "cpu",
      "memory_size_bytes": 7516192768,
      "status": "loaded"
    }
  ],
  "gpus": [
    {
      "id": 0,
      "name": "NVIDIA GeForce RTX 4090",
      "memory_total_bytes": 24576000000,
      "memory_free_bytes": 18350000000,
      "compute_capable": true
    }
  ]
}
```

### Model Stats

**Endpoint:** `GET /models/stats`

Returns registry of all registered models.

**Response Example:**
```json
{
  "status": "ok",
  "total_loaded": 2,
  "models": [
    {
      "id": "qwen-7b",
      "name": "Qwen/Qwen-7B-GGUF/qwen-7b.gguf",
      "device": "cpu",
      "memory_size_bytes": 7516192768,
      "status": "loaded"
    }
  ]
}
```

### GPU Stats

**Endpoint:** `GET /gpu/stats`

Returns available GPU hardware information.

**Response Example:**
```json
{
  "status": "ok",
  "gpus": [
    {
      "id": 0,
      "name": "NVIDIA GeForce RTX 4090",
      "memory_total_bytes": 24576000000,
      "memory_free_bytes": 18350000000,
      "compute_capable": true
    }
  ]
}
```

### Model Discovery (Opencode Integration)

**Endpoint:** `GET /models/discover`

Returns model registry information for Opencode agent discovery and integration.

**Response Example:**
```json
{
  "service_name": "llm-proxy-model-server",
  "version": "1.0.0",
  "description": "LLM Proxy model registry for agent integration and discovery",
  "model_count": 2,
  "modelsWrapped": [
    {
      "id": "qwen-7b",
      "name": "Qwen/Qwen-7B-GGUF/qwen-7b.gguf",
      "qualified_name": "Qwen/Qwen-7B-GGUF/qwen-7b.gguf",
      "device": "cpu",
      "memory_size_bytes": 7516192768,
      "status": "loaded"
    },
    {
      "id": "mistral-7b",
      "name": "Mistral/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct.Q4_K_M.gguf",
      "qualified_name": "Mistral/Mistral-7B-Instruct-v0.1-GGUF/mistral-7b-instruct.Q4_K_M.gguf",
      "device": "cpu",
      "memory_size_bytes": 6438259200,
      "status": "loaded"
    }
  ],
  "endpoint_base_url": "https://localhost:9999",
  "endpoint_path": "/models/stats"
}
```

### Model Load Endpoint

**Endpoint:** `POST /models/load`

Auto-load a model from disk path.

**Request Body:**
```json
{
  "name": "my-model",
  "path": "/path/to/model.gguf",
  "device": "cpu"
}
```

**Response:**
```json
{
  "success": true,
  "model": "my-model",
  "name": "Qwen/Qwen-7B-GGUF/qwen-7b.gguf",
  "device": "cpu",
  "url": "http://localhost:1234/v1/chat/completions",
  "status": "loaded"
}
```

### Model Proxying (Request Routing)

**Endpoint:** `GET/POST /model-{model_name}/{path}`

Proxy requests to configured backend services.

**Example Request:**
```bash
curl https://localhost:9999/model-qwen-7b/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen-7b",
    "messages": [
      {
        "role": "user",
        "content": "Hello, how are you?"
      }
    ]
  }'
```

---

## CLI Commands

### Model Management

#### List Models

```bash
# Table format (default)
./llm-proxy-manager models list

# JSON format
./llm-proxy-manager models list --format json
```

**Output Example:**
```
NAME                              DEVICE  RAM    VRAM
qwen-7b                           cpu     7401   0
mistral-7b                        cpu     6258   0
```

#### Check Model Status

```bash
./llm-proxy-manager models check qwen-7b
# Output: ✓ Model qwen-7b is loaded
```

#### Reload Model

```bash
# Reload single model
./llm-proxy-manager models reload qwen-7b

# Reload all models
./llm-proxy-manager reload --all
```

#### Unload Model

```bash
./llm-proxy-manager models unload qwen-7b
```

### Routing Configuration

#### View Routing Map

```bash
./llm-proxy-manager routing show
```

#### Manage Backends

```bash
# Add backend for model
./llm-proxy-manager backends add \
  http://localhost:1234/v1/chat/completions \
  --model qwen-7b

# Remove backend
./llm-proxy-manager backends remove \
  http://localhost:1234/v1/chat/completions
```

### Health Checks

#### Overall Proxy Health

```bash
./llm-proxy-manager health
# Output: ✓ Proxy is healthy
```

---

## Opencode Agent Integration

Opencode enables other AI agents to discover, authenticate with, and interact with models loaded in LLM Proxy.

### Setup for Agents

1. **Initialize Configuration:**
```bash
./llm-proxy-manager opencode init --proxy-url http://localhost:9999
```

2. **Generated Configuration:**
Creates `.opencode/models.yaml` with:
- Proxy URL and path settings
- API key authentication
- Rate limiting options
- Discovery endpoint configuration

3. **List Available Models:**
```bash
./llm-proxy-manager opencode list
```

### Agent Configuration Schema

```yaml
# .opencode/models.yaml
agent_name: my-agent
model_registry_url: http://localhost:9999
model_path: /models/discover
api_key: sk-your-api-key-here
rate_limit_requests_per_minute: 60
log_level: INFO
discovery_enabled: true
```

### Access Discovery Endpoint

Agents can visit the discovery endpoint to list available models:

```bash
curl https://localhost:9999/models/discover
```

Response includes:
- Service name and version
- Model count
- Array of available models with metadata
- Authentication requirements

---

## Development

### Project Structure

```
llm-proxy/
├── cmd/
│   ├── proxy/            # Proxy server main()
│   └── management/       # CLI tools main()
├── pkg/
│   ├── registry/         # Model registry & lifecycle
│   ├── memory/           # Memory pool manager
│   ├── router/           # Path-based routing
│   ├── discovery/        # LM Studio discovery
│   ├── config/           # YAML configuration loading
│   ├── hardware/         # GPU detection
│   ├── rate_limiter/     # Token bucket rate limiting
│   ├── normalizer/       # OpenAI response normalization
│   └── utils/            # Utility functions
├── config/               # Default model configurations
├── .opencode/            # Agent configuration (generated)
└── tests/                # Integration tests
```

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run specific package
go test ./pkg/registry/... -v

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building

```bash
# Build proxy server
go build -o llm-proxy-server ./cmd/proxy

# Build CLI tools
go build -o llm-proxy-manager ./cmd/management

# Build with CGO for better performance (if available)
CGO_ENABLED=1 go build -ldflags="-s -w" ./...
```

### Docker Development

```bash
docker-compose up
```

---

## Troubleshooting

### Common Issues

#### Model Won't Load

```bash
# Check memory pool
./llm-proxy-manager models list

# Reload model after changes
./llm-proxy-manager reload --all

# Check GPU availability
curl https://localhost:9999/gpu/stats
```

#### Proxy Returns Errors to Clients

- Ensure backend LM Studio/Ollama is running on expected port (default: 1234)
- Check route configuration matches model name patterns
- Verify model is loaded and status is "loaded"

```bash
# View current model status
curl https://localhost:9999/models/stats
```

#### Rate Limit Errors (429)

Increase rate limit in environment:

```bash
export RATE_LIMIT_MAX_TOKENS=500
export RATE_LIMIT_REFILL_RATE=50
./llm-proxy-server
```

**Or disable rate limiting completely:**

```bash
export DISABLE_RATE_LIMITING=true
./llm-proxy-server
```

This removes all rate limiting overhead for maximum throughput.

#### Discovery Endpoint Not Working

Ensure LM Studio discovery URL is accessible:

```bash
# Check if discovery endpoint is responding
curl http://localhost:1234/api/v1/models
```

If using custom port:

```bash
export LM_STUDIO_DISCOVERY_URL=http://localhost:1235/api/v1/models
./llm-proxy-server
```

### Debug Mode

Enable detailed logging:

```bash
export LOG_LEVEL=debug
./llm-proxy-server
```

### Reset Everything

```bash
# Unload all models
curl -X DELETE https://localhost:9999/model-qwen-7b

# Or use CLI
./llm-proxy-manager models unload qwen-7b
```

---

## License

MIT License - See LICENSE file for details.

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## Support

For issues and questions, please open an issue on GitHub or contact the maintainers.
