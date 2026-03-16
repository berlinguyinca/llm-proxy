## Phase 9: Opencode Integration - PLANNING

### Overview
Create a local registration mechanism that allows Opencode agents to discover, register with, and interact with the LLM Proxy system.

### Implementation Plan

**Wave 1: Core Integration** (2-3 hours)
1. Create `/models/discover` endpoint exposing model registry
2. Add CLI commands for Opencode configuration management  
3. Create `.opencode/models.yaml` schema
4. Implement model metadata export functionality

### Success Criteria
- ✅ CLI command to register proxy with Opencode: `llm-proxy-manager models opencode init`
- ✅ Discovery endpoint returns model registry data in Opencode-compatible format
- ✅ `.opencode/models.yaml` supports local agent configuration
