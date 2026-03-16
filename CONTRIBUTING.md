# Contributing to LLM Proxy

Thank you for your interest in contributing to LLM Proxy! This document guides you through the process.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Architecture Decisions](#architecture-decisions)
- [Common Areas for Improvement](#common-areas-for-improvement)

---

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](https://contributor-covenant.org/). By participating in this project you agree to abide by its terms.

---

## Getting Started

### Prerequisites

- Go 1.21+
- Understanding of HTTP servers and model loading concepts
- Experience with memory management (helpful but not required)

### Development Environment Setup

```bash
# Clone the repository
git clone https://github.com/your-org/llm-proxy.git
cd llm-proxy

# Download dependencies
go mod download

# Run tests to ensure everything works
go test ./...
```

### Build Commands

```bash
# Build proxy server
go build -o llm-proxy-server ./cmd/proxy

# Build CLI tools  
go build -o llm-proxy-manager ./cmd/management

# Build for production (with optimizations)
CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(git describe --tags)" ./cmd/proxy
```

---

## Development Workflow

### Branching Strategy

```
main              # Production-ready code
├── develop       # Integration branch (if using feature flags)
└── feature/*     # Feature branches from main or develop
    ├── feat/model-registration
    └── fix/memory-leak
```

**Recommended:** Create branches from `main` for all contributions.

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
feat(models): add model hot-swap capability
fix(rate-limit): prevent token bucket exhaustion
refactor(memory): improve pool allocation strategy
docs(api): update endpoint documentation
test(health): add coverage for memory threshold edge case
chore(deps): update prometheus client to v1.20.0
```

**Format:** `<type>(<scope>): <description>`

- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code reorganization with no behavior change
- `docs` - Documentation changes
- `test` - Test changes
- `chore` - Maintenance tasks

### Before Submitting PR

```bash
# Run all tests
go test ./... -v

# Check code coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Build and verify CLI tools
go build -o llm-proxy-manager ./cmd/management
./llm-proxy-manager --help

# Verify formatting (optional: install gofumpt)
gofumpt -l .

# Check for unused imports
golangci-lint run --no-config --disable-all --enable=unusedimport,godox,staticcheck
```

---

## Code Style

### General Guidelines

- Use `go fmt` to format code consistently
- Keep functions under 50 lines when possible
- Prefer composition over inheritance
- Document complex logic with comments

### Function Documentation

```go
// ModelRegistry manages model lifecycle operations
type ModelRegistry struct {
    models map[string]*ModelInfo
}

// Register adds a model to the registry
// Returns the registered model info and any error.
// If model already exists, updates its configuration.
func (r *ModelRegistry) Register(id, name string, device string, sizeBytes uint64) (*ModelInfo, error) {
    // Implementation
}
```

### Error Handling

Use `fmt.Errorf` with `%w` for wrapping errors:

```go
// GOOD: Preserves error chain
configData, err := os.ReadFile(path)
if err != nil {
    return nil, fmt.Errorf("failed to read config: %w", err)
}

// BAD: Loses error context
return nil, errors.New("config read failed")
```

### Type Safety

Always check map values before use:

```go
// GOOD: Type-safe access
func (m *ModelManager) GetDevice(modelName string) (string, bool) {
    if device, ok := m.registry.GetDevice(modelName); ok {
        return device, true
    }
    return "", false
}

// BAD: Panics on missing key
device := m.registry.GetDevice(modelName).device  // Will panic!
```

---

## Testing Guidelines

### Test Structure

Each package should have a `_test.go` file with unit tests.

```go
package registry

import (
    "testing"
)

func TestModelRegistry_Register(t *testing.T) {
    r := NewModelRegistry()
    
    entry, err := r.Register("test", "model.gguf", "cpu", 1024*1024*1024)
    
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if entry.ID != "test" {
        t.Errorf("Expected ID 'test', got '%s'", entry.ID)
    }
}
```

### Testing Strategies

1. **Unit Tests**: Test individual functions in isolation
2. **Integration Tests**: Test complete workflows (e.g., proxy request → backend)
3. **Fuzz Tests**: Find input vulnerabilities (use Go's built-in fuzzing)
4. **Benchmark Tests**: Measure performance improvements

### Example: Comprehensive Unit Test

```go
func TestMemoryPoolManager_MemoryCalculation(t *testing.T) {
    pm := memory.NewMemoryPoolManager(10*1024*1024*1024, 5) // 10GB pool, 5GB threshold
    
    // Add model under threshold
    err := pm.AddModel("model-1", 4*1024*1024*1024, "cpu")
    if err != nil {
        t.Fatalf("Failed to add model: %v", err)
    }
    
    // Verify pool state
    pools := pm.GetAllPools()
    if len(pools) != 2 {
        t.Errorf("Expected 2 pools, got %d", len(pools))
    }
}
```

### Coverage Target

Maintain >80% test coverage for core packages:
- `pkg/registry` - Model lifecycle
- `pkg/memory` - Memory management  
- `pkg/rate_limiter` - Rate limiting logic
- `pkg/router` - Routing logic

---

## Pull Request Process

### PR Requirements

1. **Title**: Clear description of change
2. **Description**: 
   - What problem does this solve?
   - How was it solved?
   - Testing performed
   - Breaking changes (if any)
3. **Tests**: New tests for new functionality, existing tests pass
4. **Documentation**: Updated README/API docs as needed
5. **Code Review**: Minimum 1 approval required

### PR Template

```markdown
## Description

Brief description of the change and problem it solves.

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing

```bash
# Commands you ran to test this change
go test ./...
./llm-proxy-manager models list
curl https://localhost:9999/health
```

Expected results:

```
✓ All tests pass
✓ CLI tools work as expected
✓ Proxy responds correctly
```

## Checklist

- [ ] My code follows the style guide
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] All tests pass with `go test ./...`

## Related Issues

Closes #[issue-number]
References #[related-issue]
