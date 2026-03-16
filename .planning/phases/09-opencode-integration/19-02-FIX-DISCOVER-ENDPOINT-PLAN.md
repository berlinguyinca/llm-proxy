---
phase: 09-opencode-integration
plan: 02
type: execute
wave: 1
depends_on:
  - 19-01
files_modified:
  - cmd/proxy/main.go
autonomous: true

must_haves:
  truths:
    - "GET /models/discover returns valid JSON with model registry data"
    - "Models array contains wrapped ModelInfo objects (not raw registry.ModelInfo)"
    - "Response includes service_name, version, description, endpoint metadata"
    - "Discovery endpoint works correctly for Opencode agent registration"
  artifacts:
    - path: "cmd/proxy/main.go"
      provides: "/models/discover handler with fixed model wrapping"
  key_links:
    - from: "discoverModelsHandler()"
      to: "GET /models/discover"
      via: "HTTP GET request handling"
      pattern: "json.MarshalIndent(discoveryResponse)"

<objective>
Fix the model wrapping bug in discoverModelsHandler() and verify the endpoint returns correct wrapped ModelInfo objects for Opencode agent discovery.

Purpose: Correctly wrap registry.ModelInfo with local ModelInfo type so discovery endpoint exposes all model metadata to external agents
Output: Fixed /models/discover handler returning properly structured discovery response
</objective>

<execution_context>
@/Users/wohlgemuth/.config/opencode/get-shit-done/workflows/execute-plan.md
@/Users/wohlgemuth/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
# Discovered Bug

In discoverModelsHandler() (lines 696-745), the code:
1. Creates `modelsWrapped` with wrapped ModelInfo objects  
2. Returns `models` instead of `modelsWrapped` in discoveryResponse

This means agents receive raw registry.ModelInfo instead of the local type, breaking Opencode integration.

# Existing Code Context

From cmd/proxy/main.go (lines 696-745):
```go
func discoverModelsHandler(manager *ModelManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        models := manager.GetModels()
    
    modelsWrapped := make([]ModelInfo, len(models))
    for i, m := range models {
        modelsWrapped[i] = ModelInfo{
            ID:            m.ID,
            Name:          m.Name,
            QualifiedName: m.QualifiedName,
            Device:        m.Device,
            MemorySize:    m.MemorySize,
            Status:        registry.Status(m.Status),
        }
    }

        discoveryResponse := struct {
            ServiceName     string              `json:"service_name"`
            Version         string              `json:"version"`
            Description     string              `json:"description"`
            ModelCount      int                 `json:"model_count"`
            Models          []ModelInfo         `json:"models"`
            EndpointBaseURL string              `json:"endpoint_base_url"`
            EndpointPath    string              `json:"endpoint_path"`
        }{
            ServiceName:     "llm-proxy",
            Version:         "1.0.0",
            Description:     "LLM Proxy model registry for agent integration and discovery",
            ModelCount:      len(modelsWrapped),
            Models:          modelsWrapped,
            EndpointBaseURL: fmt.Sprintf("http://%s", r.Host),
            EndpointPath:    "/models/stats",
        }

        responseBytes, err := json.MarshalIndent(discoveryResponse, "", "  ")
        if err != nil {
            log.Printf("Error marshaling discovery response: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        if _, writeErr := w.Write(responseBytes); writeErr != nil {
            log.Printf("Error writing response: %v", writeErr)
        }
    }
}
```

# Required Changes

The discoverModelsHandler has been fixed to:
- Use `modelsWrapped` instead of `models` in the response struct
- Removed duplicate DiscoveryEndpoint type definition
- Added model_count and endpoint fields for better agent discovery
</context>

<tasks>

<task type="auto">
  <name>Create .opencode directory with models.yaml.example schema</name>
  <files>.opencode/models.yaml.example</files>
  <action>
    Create the Opencode configuration directory and example file with:
    
    1. YAML Schema documenting the configuration format for local agent registration
       - Required fields: proxy_url, proxy_path, authentication (api_key or bearer_token)
       - Optional fields: models (list of registered model names), rate_limit_tokens
    
    2. Example configuration showing how an Opencode agent should configure itself to work with LLM Proxy
    
    3. Include comments explaining each field and its purpose for agent integration
    
    This enables Opencode agents to discover, authenticate, and interact with loaded models in LLM Proxy.
  </action>
  <verify>ls -la .opencode/ && cat .opencode/models.yaml.example | head -50</verify>
  <done>.opencode directory created with models.yaml.example containing complete schema documentation and example configuration for Opencode agent integration</done>
</task>

<task type="auto">
  <name>Add Opencode CLI commands to management tool</name>
  <files>cmd/management/main.go</files>
  <action>
    Add two new CLI subcommands under `llm-proxy-manager models opencode`:
    
    1. "init" subcommand:
       - Creates .opencode/models.yaml configuration file for local agent registration
       - Accepts optional --proxy-url flag (defaults to http://localhost:9999)
       - Generates template with API key placeholder
       - Prints setup instructions after creation
    
    2. "list" subcommand:
       - Reads .opencode/models.yaml if it exists
       - Displays registered proxy configuration and available models
       - Shows authentication method (API key or bearer token)
    
    Add these commands to the root command registration in main() function, following existing Cobra command patterns.
  </action>
  <verify>go run ./cmd/management/main.go --help</verify>
  <done>New CLI subcommands available: llm-proxy-manager models opencode init and llm-proxy-manager models opencode list</done>
</task>

<task type="checkpoint:human-verify" gate="blocking">
  <what-built>Fixed /models/discover endpoint and added Opencode integration (CLI + config schema)</what-built>
  <how-to-verify>
    1. Start the proxy server: go run ./cmd/proxy/main.go > /dev/null 2>&1 &
    
    2. Test discovery endpoint:
       curl -s http://localhost:9999/models/discover | jq .
    
    3. Verify response structure includes:
       - service_name: "llm-proxy"
       - version: "1.0.0"
       - description field
       - model_count (integer)
       - endpoint_base_url and endpoint_path
       - models array with wrapped ModelInfo objects
    
    4. Test CLI commands (if proxy is running):
       go run ./cmd/management/main.go models opencode init --help
       go run ./cmd/management/main.go models opencode list --help
    
    5. Verify .opencode directory structure:
       ls -la .opencode/
    
    6. Report back with actual output from these commands
  </how-to-verify>
  <resume-signal>Type "approved" when verification is complete, or describe any issues</resume-signal>
</task>

</tasks>

<verification>
Run: go run ./cmd/proxy/main.go & (then) curl http://localhost:9999/models/discover
</verification>

<success_criteria>
- GET /models/discover returns HTTP 200 with valid JSON containing wrapped ModelInfo objects
- Response includes all metadata fields: service_name, version, description, model_count, endpoint_*
- .opencode directory exists with models.yaml.example file
- CLI commands "llm-proxy-manager models opencode init" and "list" are available
</success_criteria>

<output>
After completion, create `.planning/phases/09-opencode-integration/19-02-FIX-DISCOVER-ENDPOINT-SUMMARY.md`
</output>
