## Phase 8: Auto-Loading Feature - PLANNING

### Overview
Implement automatic model loading when a user requests a model that isn't currently loaded. This enhances the LLM Proxy by allowing seamless model loading without manual intervention, improving the developer experience and enabling dynamic workloads.

### Requirements Analysis

**Current Architecture:**
1. Models are registered in `pkg/registry/` with status tracking (unloaded/loading/loaded)
2. Requests go through `/model-` prefix routing to configured backend URLs
3. Proxy handles requests via `/chat/completions` endpoint in router

**Auto-Loading Feature Requirements:**
1. **Endpoint**: Add `/models/load` HTTP endpoint that accepts model name and loads it on-demand
2. **Routing Enhancement**: Check model status before proxying, auto-load if needed
3. **Status Tracking**: Persist model load state across requests (memory-based for now)
4. **Error Handling**: Handle load failures gracefully with appropriate error messages

### Implementation Strategy

**Approach 1: Auto-Load Hook in Router** (Recommended)
- Add pre-proxy hook that checks if target backend's model is loaded
- If not loaded, call `/models/load` endpoint before proxying
- Simpler architecture, clear separation of concerns

**Approach 2: Proxy Middleware Layer**
- Create middleware layer that sits between router and handler
- Check/load models in middleware before routing
- More flexible but adds complexity

**Decision**: Use Approach 1 (Auto-Load Hook) for initial implementation

### Proposed API Design

**New Endpoint:**
```
POST /models/load
Body: {
  "name": "qwen2.5-7b-chat",
  "path": "/data/models/qwen.gguf"
}
Response: {
  "success": true,
  "message": "Model loaded successfully",
  "model_name": "qwen2.5-7b-chat"
}
```

**Auto-Load Configuration:**
```yaml
# models.yaml or environment variable
auto_load_models:
  enabled: true
  timeout_ms: 30000  # Maximum time to load a model
  on_demand_only: false  # Don't load if not configured in registry first
```

### Wave Planning

**Wave 1: Core Implementation** (~4-6 hours)
1. Create `/models/load` endpoint with registration & loading logic
2. Add auto-load hook in router.ServeHTTP()
3. Implement model status persistence (in-memory for now)
4. Add unit tests for load/unload operations
5. Add integration test for auto-load behavior

**Wave 2: Enhancement & Documentation** (~2-3 hours)
1. Add streaming progress indication during model load
2. Add metrics endpoint for model load statistics
3. Update documentation with usage examples
4. Add error handling and timeout configuration

### Files to Modify/Create

| File | Type | Purpose |
|------|------|---------|
| `cmd/proxy/main.go` | modify | Add `/models/load` HTTP handler |
| `pkg/registry/manager.go` | modify | Add LoadFromRegistry() method |
| `pkg/router/router.go` | modify | Add auto-load hook before proxying |
| `.env.example` | modify | Add AUTO_LOAD config options |
| `cmd/proxy/main_test.go` | add tests | Unit tests for load operations |

### Success Criteria

- ✅ Models can be loaded on-demand via API endpoint
- ✅ Auto-loading triggers when unregistered model requested
- ✅ Load failures are handled gracefully
- ✅ All existing endpoints continue working
- ✅ Code coverage maintained at 90%+
