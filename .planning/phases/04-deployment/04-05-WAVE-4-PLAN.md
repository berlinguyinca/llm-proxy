---
phase: 01-foundation
plan: 05
type: execute
wave: 1
depends_on: []
files_modified: 
  - "Dockerfile"
  - "docker-compose.yml"
  - "cmd/proxy/metrics.go"
  - ".env.example"
  - "config/models.yaml"
autonomous: true

must_haves:
  truths:
    - "Proxy container builds successfully in production environment"
    - "All health endpoints accessible via Docker networking"
    - "Prometheus metrics available at /metrics endpoint"
    - "Docker Compose up-and-running with proxy service"
    - ".env.example contains all required deployment configuration variables"
  artifacts:
    - path: "Dockerfile"
      provides: "Production-optimized Go binary builder image"
      min_lines: 30
    - path: "docker-compose.yml"
      provides: "Multi-service deployment orchestration"
      exports: ["proxy service", "lm-studio service (optional)"]
    - path: "cmd/proxy/metrics.go"
      provides: "Prometheus metrics instrumentation for /metrics endpoint"
      contains: ["prometheus", "registry", "counter", "histogram"]
  key_links:
    - from: "Dockerfile"
      to: "go mod dependencies"
      via: "multi-stage build copying only production binaries"
      pattern: "COPY --from=builder ..."
    - from: "docker-compose.yml"
      to: "proxy service"
      via: "image reference and port mapping"
      pattern: "build:" or "image:" for proxy
    - from: "cmd/proxy/metrics.go"
      to: "http.Handle"
      via: "registering metrics handlers"
      pattern: "http\.Handle\(\"/metrics\""

---

## Objective

Complete Wave 4: Production Deployment & Monitoring by adding Dockerfile, docker-compose.yml, Prometheus metrics instrumentation, and comprehensive deployment documentation.

**Purpose**: Enable production-ready deployment with containerization, service orchestration, observability, and streamlined operational procedures.

**Output**: 
- `Dockerfile` - Multi-stage build for production binary
- `docker-compose.yml` - Service orchestration with optional LM Studio
- `cmd/proxy/metrics.go` - Prometheus metrics instrumentation
- Updated `.env.example` with deployment variables
- Updated `config/models.yaml` with resource hints

## Execution Context

From previous waves:
- @.planning/ROADMAP.md
- @.planning/STATE.md
- @cmd/proxy/main.go
- @pkg/rate_limiter/token_bucket.go
- @go.mod

## Tasks

<task type="auto">
  <name>Task 1: Create production Dockerfile with multi-stage build</name>
  <files>Dockerfile</files>
  <action>
    Create multi-stage Dockerfile for optimized production binary:
    
    **Stage 1 (Builder)**:
    - Use `golang:1.22-alpine` as base image
    - Install Go 1.22, ca-certificates, tzdata
    - WORKDIR /build
    - COPY go.mod and go.sum first for cache optimization
    - RUN `go mod download` to populate module cache
    - COPY all source files (./*) into /build
    - RUN `CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build` with flags:
      - `-ldflags="-s -w -X main.Version=$(git describe --tags 2>/dev/null || echo 'v0.0.0')"`
    
    **Stage 2 (Production)**:
    - Use `gcr.io/distroless/static-debian11` as production base
    - WORKDIR /app
    - COPY --from=builder /build/bin/proxy ./proxy
    - USER non-root user
    
    **Image size target**: ~20-35MB (vs ~900MB for full Go SDK image)
    
  </action>
  <verify>
    docker build -t llm-proxy:dev -f Dockerfile .
    docker images | grep llm-proxy
    
    Expected output:
    - IMAGE         SIZE        CREATED
    - llm-proxy:dev  ~30M        <seconds ago>
  </verify>
  <done>
    Dockerfile exists with multi-stage build (2 stages)
    Stage 1 uses golang:1.22-alpine for building
    Stage 2 uses distroless/static-debian11 for production
    CGO_ENABLED=1 set for GPU runtime compatibility
    -ldflags includes version string from git tags
    File ends with USER directive for non-root execution
  </done>
</task>

<task type="auto">
  <name>Task 2: Create docker-compose.yml with proxy and optional LM Studio services</name>
  <files>docker-compose.yml</files>
  <action>
    Create docker-compose.yml with the following configuration:
    
    **Proxy Service**:
    - image: llm-proxy or use build context
    - ports: ["9999:9999"] - proxy exposed to host
    - environment variables from .env file
    - volumes: ./config:/app/config (models.yaml mounted)
    - healthcheck with curl to /health endpoint
    - restart: unless-stopped
    
    **LM Studio Service (optional, commented out)**:
    - image: lmstudio/lmstudio:latest
    - ports: ["1234:1234"]
    - volumes for model persistence if needed
    
    **Network Configuration**:
    - internal network for proxy <-> lm-studio communication
    
  </action>
  <verify>
    docker-compose config
    
    Expected output:
    - Validates YAML syntax
    - Shows resolved image names or build contexts
    - Lists all environment variables
    - Displays healthcheck configuration
  </verify>
  <done>
    docker-compose.yml file exists with valid YAML syntax
    Proxy service defined with port mapping 9999:9999
    All environment variables from .env.example documented
    Healthcheck configured for proxy container
    LM Studio service commented out (optional for development)
    Network configuration enables proxy-lm-studio communication
  </done>
</task>

<task type="auto">
  <name>Task 3: Add Prometheus metrics instrumentation to proxy</name>
  <files>cmd/proxy/metrics.go</files>
  <action>
    Create new file cmd/proxy/metrics.go with Prometheus metrics:
    
    **Imports**:
    - github.com/prometheus/client_golang/prometheus
    - github.com/prometheus/client_golang/prometheus/promhttp
    
    **Metrics Defined**:
    1. Request counter (CounterVec) for requests by status and model
    2. Latency histogram (HistogramVec) for request duration by model
    3. Active connections gauge for monitoring
    
    **Create HTTP handler for /metrics endpoint**
    
    **Register metrics on startup** in main.go
    
  </action>
  <verify>
    go build -o bin/proxy ./cmd/proxy 2>&1
    
    Expected output:
    - No compilation errors
    - Binary built successfully (~9-10MB)
  </verify>
  <done>
    cmd/proxy/metrics.go file created with Prometheus packages imported
    CounterVec for requestsTotal defined with status and model labels
    HistogramVec for requestDuration defined with model label
    metricsHandler() function returning promhttp.Handler()
    http.Handle("/metrics", ...) registration in main.go or constructor
    MustRegister calls for all metrics added
  </done>
</task>

<task type="auto">
  <name>Task 4: Update .env.example with deployment variables and documentation</name>
  <files>.env.example</files>
  <action>
    Append deployment settings to .env.example:
    
    PORT=9999
    MEMORY_THRESHOLD_GB=16
    RATE_LIMIT_MAX_TOKENS=100
    RATE_LIMIT_REFILL_RATE=10
    DISCOVERY_ENABLED=false
    LOG_LEVEL=info
    
    Add documentation section explaining rate limiting and metrics at end of file
    
  </action>
  <verify>
    grep -E "(PORT=|RATE_LIMIT|DISCOVERY_ENABLED|LOG_LEVEL)" .env.example
    
    Expected output:
    Shows at least 5 environment variable entries
  </verify>
  <done>
    .env.example file contains all deployment variables
    RATE_LIMIT_MAX_TOKENS documented with purpose
    DISCOVERY_ENABLED setting added
    LOG_LEVEL configuration added
    Comments section at end explaining rate limiting and metrics
  </done>
</task>

<task type="auto">
  <name>Task 5: Add resource hints to models.yaml</name>
  <files>config/models.yaml</files>
  <action>
    Update config/models.yaml to add optional resource hints field:
    
    min_memory_mb: 4500       # Minimum memory required (optional)
    eviction_priority: 1      # Eviction priority when pool is full (1-9)
    
  </action>
  <verify>
    cat config/models.yaml | grep -E "(min_memory_mb|eviction_priority)"
    
    Expected output:
    Shows models with resource hint fields
  </verify>
  <done>
    config/models.yaml updated with optional resource hints schema
    min_memory_mb field documented as optional hint
    eviction_priority field added for LRU management
    File still valid YAML and builds without errors
  </done>
</task>

<verification>
1. **Build verification**: docker build -t llm-proxy:test -f Dockerfile .
2. **Compose configuration check**: docker-compose config --quiet
3. **Code compilation**: go build -o bin/proxy ./cmd/proxy && echo "Build: OK"
4. **Unit tests pass**: go test ./... -timeout 60s
</verification>

<success_criteria>
- Dockerfile exists and validates with docker build
- docker-compose.yml is syntactically valid
- Prometheus metrics code compiles without errors
- All health endpoints still work after changes
- Binary size acceptable (~9-10MB)
- Documentation updated with deployment procedures
</success_criteria>

<output>
After completion, create .planning/phases/04-deployment/{phase}-05-SUMMARY.md
</output>
</content>