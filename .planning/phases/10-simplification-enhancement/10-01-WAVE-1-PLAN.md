---
phase: 10-simplification-enhancement
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: 
  - cmd/management/main.go
  - docs/admin-guide.md
  - pkg/health/endpoints.go
  - cmd/proxy/main.go
autonomous: true

must_haves:
  truths:
    - "Opencode CLI commands work without nesting (llm-proxy-manager opencode init instead of llm-proxy-manager llm-proxy-manager opencode init)"
    - "Admin guide covers model deployment, scaling, and operational procedures"
    - "Health endpoint shows memory pressure with actionable thresholds"
    - "Auto-load feature automatically loads models based on available resources"
    - "Opencode integration allows agents to discover and register with proxy"
  artifacts:
    - path: cmd/management/main.go
      provides: "Fixed CLI command structure without nesting"
      contains: "opencodeInitCmd, opencodeListCmd as root-level commands"
    - path: docs/admin-guide.md
      provides: "Complete system administrator procedures"
      min_lines: 500
    - path: pkg/health/endpoints.go
      provides: "Enhanced health metrics with memory pressure"
      contains: "/health endpoint with memory_utilization, swap_pressure fields"
    - path: cmd/proxy/main.go
      provides: "Auto-load initialization logic"
      exports: ["NewModelManager"]
    - path: .opencode/models.yaml.example
      provides: "Opencode configuration template"
      min_lines: 20
  key_links:
    - from: "cmd/management/main.go"
      to: "pkg/model_manager.go"
      via: "CLI commands wire to model manager operations"
      pattern: "modelManager.ReloadModels\\(\\)"
    - from: "pkg/health/endpoints.go"
      to: "cmd/proxy/main.go"
      via: "health endpoint registered in HTTP server"
      pattern: "http.Handle.*\\/health"
---

<objective>
Complete all remaining simplification items identified in Agent Analysis: fix CLI command nesting, create admin guide, enhance health metrics, implement auto-load feature, and complete opencode integration.

Purpose: Remove the last 23 complexities identified through agent persona analysis, making the proxy production-ready for new developers, DevOps engineers, Opencode agents, and system administrators.

Output: Working CLI without command nesting errors, comprehensive admin documentation, enhanced monitoring, automatic resource-based model loading, full Opencode agent integration.
</objective>

<context>
@/Users/wohlgemuth/IdeaProjects/llm-proxy/docs/AGENT-ANALYSIS.md
@/Users/wohlgemuth/.planning/phases/09-opencode-integration/PLANNING_SUMMARY.md

# Key Context from Previous Phases

**Phase 1 Foundation (Complete):** Core proxy infrastructure with router, registry, memory manager, hardware detection, health endpoints. Binary builds at ~9MB with ~75%+ test coverage.

**Phase 6 CLI Management (Complete):** llm-proxy-manager CLI for model list/reload/unload operations. All commands tested and verified working.

**Phase 8 Auto-Load Feature:** Model manager initialized with memory monitor that tracks usage and evicts models when thresholds exceeded (auto-load logic skeleton exists in pkg/model_manager.go).

**Phase 9 Opencode Integration:** Discovery endpoint created at /models/discover, CLI init command works but has nesting issue (lines 679-782 in main.go show duplicate registration pattern).
</context>

<tasks>

<task type="auto">
  <name>Task 0: Fix opencode CLI command nesting - remove duplicate registration pattern</name>
  <files>cmd/management/main.go</files>
  <behavior>
    Case 1: When "llm-proxy-manager" subcommand exists at lines ~679, expect nested command structure error
    Case 2: After fix, expect single-level commands like "llm-proxy-manager opencode init" to work
    Case 3: Verify all opencode commands (init, list) are registered once, not twice
  </behavior>
  <action>
    Analyze lines 679-782 in cmd/management/main.go to find duplicate command registration pattern. The issue is that opencodeInitCmd() and opencodeListCmd() are being registered under both a top-level "llm-proxy-manager" command AND directly with root commands, causing "opencode opencode init" nesting error.
    
    Fix approach:
    1. Remove the nested "llm-proxy-manager" registration block for opencode commands (the duplicate pattern)
    2. Keep only direct registration of opencode commands with root cobra.Command instance
    3. Update command structure to have flat hierarchy: llm-proxy-manager models | routing | opencode (no double-nesting)
    4. Verify all subcommands are registered exactly once
    
    After fix, CLI invocations should be:
      - llm-proxy-manager models list ✅ (was working)
      - llm-proxy-manager models reload ✅ (was working)  
      - llm-proxy-manager opencode init ✅ (currently shows nesting error after double-registration)
      - llm-proxy-manager routing inspect ✅ (was working)
  </action>
  <verify>
    <automated>go build -o bin/llm-proxy-manager ./cmd/management && bin/llm-proxy-manager opencode init --help 2>&1 | grep -q "Initialize Opencode" && echo "CLI nesting fixed"</automated>
  </verify>
  <done>Opencode CLI commands work without nested registration; help output shows correct command structure</done>
</task>

<task type="auto">
  <name>Task 1: Create comprehensive admin guide with deployment procedures</name>
  <files>docs/admin-guide.md</files>
  <action>
    Create complete admin guide covering:
    
    Chapter 1: Quick Reference
      - Common commands (llm-proxy-manager models list/reload/unload)
      - Rate limiting configuration and troubleshooting  
      - Docker deployment checklist
    
    Chapter 2: Model Deployment
      - Loading and managing multiple models
      - Memory threshold configuration
      - GPU vs CPU placement decisions
      - Eviction priority settings in models.yaml
    
    Chapter 3: Scaling Operations
      - Horizontal scaling with multiple proxy instances
      - Shared memory pool setup between instances
      - Load balancer configuration
      - Health check endpoints for orchestrators
    
    Chapter 4: Monitoring & Alerting
      - Prometheus metrics explanation and queries
      - Grafana dashboard integration
      - Alert rule examples (memory pressure, rate limit exhaustion)
    
    Chapter 5: Troubleshooting Guide
      - Common error messages and fixes
      - GPU detection issues
      - Model loading failures
      - Rate limiter tuning guides
    
    Format: Markdown with code blocks, tables for command references, numbered procedures. Minimum 500 lines of content.
  </action>
  <verify>
    <automated>wc -l docs/admin-guide.md && head -50 docs/admin-guide.md | grep -q "Common Commands" && echo "Admin guide created with proper structure"</automated>
  </verify>
  <done>Comprehensive admin guide exists at docs/admin-guide.md covering model deployment, scaling, monitoring, and troubleshooting procedures</done>
</task>

<task type="auto">
  <name>Task 2: Enhance /health endpoint with memory pressure metrics</name>
  <files>pkg/health/endpoints.go, cmd/proxy/main.go</files>
  <behavior>
    Case 1: Current /health shows only basic status (OK/not OK)
    Case 2: After enhancement, must show: memory_utilization, swap_pressure, gpu_memory_utilization, evictions_pending
    Case 3: Response must be JSON format with all metrics at root level
  </behavior>
  <action>
    Modify pkg/health/endpoints.go /health handler to include:
    
    Existing (keep):
      - Status field (string)
      - Uptime field (int64)
      - Models_loaded count
      - Total_memory_mb field
    
    Add new fields:
      - memory_utilization: float percentage of pool utilization (e.g., 73.5%)
      - swap_pressure: boolean flag when system uses swap actively
      - gpu_memory_utilization: optional float if GPU available
      - evictions_pending: int count of models queued for eviction
    
    Wire health metrics to existing MemoryMonitor in cmd/proxy/main.go:
      - memoryMonitor.MemoryUtilization() -> response
      - memoryMonitor.SwapPressure() -> response  
      - memoryMonitor.EvictionPendingCount() -> response
    
    Keep /health endpoint lightweight (still suitable for load balancer probes) but provide actionable metrics.
  </action>
  <verify>
    <automated># Start proxy in background, then curl and verify new fields
    go build -o bin/proxy ./cmd/proxy &
    sleep 3
    response=$(curl -s http://localhost:9999/health)
    echo "$response" | jq -r '.memory_utilization' && echo "Memory utilization field present"</automated>
  </verify>
  <done>/health endpoint includes memory_utilization, swap_pressure, gpu_memory_utilization, evictions_pending fields</done>
</task>

<task type="auto">
  <name>Task 3: Implement auto-load feature in model manager initialization</name>
  <files>cmd/proxy/main.go, pkg/model_manager.go</files>
  <behavior>
    Case 1: Before init, models load on-demand when first request arrives
    Case 2: After init, models with resource hints load automatically if resources available
    Case 3: Auto-load checks: enough memory free (min_memory_mb threshold), GPU VRAM available
  </behavior>
  <action>
    Enhance model manager initialization in cmd/proxy/main.go to auto-load based on resource hints:
    
    In NewModelManager() after config loading:
      - Iterate through loaded models from config/models.yaml
      - For each model with "min_memory_mb" hint:
          * Check if current memory pool utilization < (100 - min_memory_mb)*percentile
          * If resources available, call manager.LoadModel() immediately
          * Skip if another model already uses that GPU device
    
    Create autoLoadModels() helper function that:
      - Takes MemoryMonitor as argument
      - Evaluates resource thresholds from each model's metadata
      - Loads models in order of eviction_priority (highest priority first)
      - Respects concurrent load limits (max 2 simultaneous loads)
    
    Add configuration options:
      - ENABLE_AUTO_LOAD: boolean (default true)
      - AUTO_LOAD_THRESHOLD_PCT: float (default 80 - load when utilization below this)
      - MAX_CONCURRENT_LOADS: int (default 2)
    
    Ensure auto-load respects:
      - Memory pool capacity
      - GPU device availability
      - Model memory requirements from config
  </action>
  <verify>
    <automated># Verify auto-load configuration is read
    response=$(curl -s http://localhost:9999/models/stats)
    echo "$response" | jq -r '.models[] | select(.name=="qwen2.5-1.5B") | .status' && echo "Auto-loaded models show status"</automated>
  </verify>
  <done>Model manager auto-loads models on startup based on resource hints and available memory</done>
</task>

<task type="auto">
  <name>Task 4: Create Opencode models.yaml.example configuration template</name>
  <files>.opencode/models.yaml.example</files>
  <action>
    Create comprehensive example configuration at .opencode/models.yaml.example:
    
    Include documented sections for:
      - proxy_url: LLM Proxy server endpoint
      - proxy_path: API path prefix (default "/")
      - authentication: api_key type with placeholder value
      - models: array placeholder with documented fields (name, path, memory requirements)
      - rate_limit: optional section with tokens and refill_rate
      - logging: level configuration
      - discovery: endpoint for model registry
    
    Add YAML comments explaining:
      - When each field is required vs optional
      - Default values if not specified
      - Examples of model entries to load automatically
      - How agents discover available models
    
    Also create .opencode/config.example as simpler YAML-instructed template.
  </action>
  <verify>
    <automated>ls -la .opencode/ && wc -l .opencode/models.yaml.example && head -30 .opencode/models.yaml.example | grep -q "# LLM Proxy" && echo "Opencode config template created"</automated>
  </verify>
  <done>Opencode configuration templates exist with complete documentation and examples</done>
</task>

<task type="auto">
  <name>Task 5: Enhance discovery endpoint with model metadata export</name>
  <files>pkg/discovery/lmstudio_api.go, cmd/proxy/main.go</files>
  <behavior>
    Case 1: Before, /models/discover returns basic list from LM Studio registry
    Case 2: After enhancement, also includes memory requirements, eviction_priority hints
    Case 3: Response format includes both discovered models and loaded models (distinguished)
  </behavior>
  <action>
    Enhance the discovery endpoint to provide richer model metadata:
    
    Modify /models/discover response structure:
      - Add fields: memory_mb_required, gpu_device_hint, eviction_priority, auto_loadable
      - Distinguish between "discovered" (from LM Studio registry) and "loaded" states
      - Include resource hints from config/models.yaml when available
    
    Wire to existing ModelManager.GetModelMetadata():
      - Return both discovery data and loaded status
      - Export resource hints alongside API details
    
    Keep endpoint backward compatible:
      - Existing clients that only read basic fields still work
      - New Opencode agents can utilize full metadata
  </action>
  <verify>
    <automated>response=$(curl -s http://localhost:9999/models/discover) && echo "$response" | jq '.models[] | select(.name=="qwen2.5-1.5B") | .memory_mb_required' && echo "Metadata enrichment working"</automated>
  </verify>
  <done>Discovery endpoint returns enriched model metadata including resource hints and load status</done>
</task>

<task type="checkpoint:human-verify">
  <what-built>All simplification fixes: CLI nesting removed, admin guide created, health metrics enhanced, auto-load implemented, Opencode integration completed</what-built>
  <how-to-verify>
    Step 1: Build and test fixed CLI
      - Run: go build -o bin/llm-proxy-manager ./cmd/management
      - Verify: llm-proxy-manager opencode init --help shows correct usage (no "opencode opencode" nesting)
      - Expected output starts with "Create .opencode/models.yaml configuration"
    
    Step 2: Verify admin guide exists and is comprehensive
      - Run: wc -l docs/admin-guide.md
      - Should show 500+ lines
      - Check sections exist: look for "Common Commands", "Model Deployment", "Scaling Operations"
    
    Step 3: Test enhanced health endpoint
      - Start proxy: go run ./cmd/proxy &
      - Wait 10 seconds for startup
      - Run: curl http://localhost:9999/health | jq .
      - Verify response includes memory_utilization, swap_pressure fields
      - Values should be numeric (e.g., "memory_utilization": 0.73 or similar)
    
    Step 4: Check auto-load is working
      - Start proxy with ENABLE_AUTO_LOAD=true (default)
      - Wait for startup complete
      - Run: curl http://localhost:9999/models/stats | jq '.models'
      - Models with resource hints should show "status": "loaded" on startup
    
    Step 5: Verify Opencode integration
      - Run: llm-proxy-manager opencode init --proxy-url http://localhost:9999
      - Verify .opencode/models.yaml created with correct proxy_url and api_key
      - Run: llm-proxy-manager opencode list shows configuration
    </how-to-verify>
  <resume-signal>Type "approved" when all manual checks pass, or describe any issues found</resume-signal>
</task>

<verification>
Run through the complete Agent Analysis checklist to verify all 18 problems are addressed:
- ✅ docker-compose.yml syntax (fixed in Phase 1)
- ✅ .env.example comprehensive (created in Phase 1)
- ✅ CLI command nesting (fixed in this plan Task 0)
- ✅ Health endpoint metrics (enhanced in this plan Task 2)
- ✅ Rate limiting optional (done in Phase 1)
- ✅ Admin guide created (Task 1)
- ✅ Auto-load implemented (Task 3)
- ✅ Opencode integration complete (Tasks 4,5)
</verification>

<success_criteria>
Phase 10 Wave 1 Complete when:
1. All CLI commands work without nesting errors (verified with llm-proxy-manager opencode init --help)
2. Admin guide exists at docs/admin-guide.md with 500+ lines covering deployment procedures
3. /health endpoint returns JSON with memory_utilization, swap_pressure, gpu_memory_utilization fields
4. Models auto-load on startup when resource hints present and memory available (verified via /models/stats)
5. Opencode config templates created at .opencode/models.yaml.example with full documentation
6. Discovery endpoint returns enriched metadata including resource hints
7. All new code compiles and tests pass (rate limiter fuzz tests still passing)
8. No breaking changes to existing API contracts
</success_criteria>

<output>
After completion, create `.planning/phases/10-simplification-enhancement/21-10-WAVE-1-SUMMARY.md`
</output>
