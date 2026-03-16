# Wave 1: Testing & Validation Goals

## Test Plan Overview

This wave focuses on comprehensive testing of all core components before proceeding to production hardening.

### Test Categories

1. **Unit Tests** - Individual component tests (<30min each)
2. **Integration Tests** - End-to-end flow tests (LM Studio connectivity)
3. **Edge Case Tests** - Error handling, boundary conditions
4. **Performance Tests** - Memory management under load scenarios

### Expected Outcomes

- All components have at least one passing unit test
- Edge cases covered (missing GPUs, invalid configs, streaming errors)
- Integration tests verify full proxy flow works end-to-end
- Code coverage targets 80%+ for core business logic

---

## Test Files to Create

### Unit Tests (pkg/*_test.go files)

1. `pkg/router/router_test.go` - Routing logic
2. `pkg/registry/manager_test.go` - Model operations
3. `pkg/memory/pool_manager_test.go` - Memory threshold enforcement
4. `pkg/hardware/detect_test.go` - GPU detection edge cases
5. `pkg/discovery/lmstudio_api_test.go` - Discovery parsing
6. `pkg/normalizer/openai_compat_test.go` - Response normalization

### Integration Tests (integration/*_test.go)

1. `integration/proxy_flow_test.go` - Full proxy request flow
2. `integration/model_loading_test.go` - Model load/unload cycles
3. `integration/memory_management_test.go` - Eviction scenarios
4. `integration/discovery_test.go` - Auto-discovery verification

### E2E Tests (cmd/*_test.go)

1. `cmd/proxy/e2e_test.go` - Full system e2e tests

---

## Validation Checklist

After all tests are written and passing:

- [ ] All unit tests pass
- [ ] Integration tests pass with mocked LM Studio
- [ ] E2E tests pass against real LM Studio (or skip if not available)
- [ ] Code coverage report generated (80%+ target)
- [ ] Documentation updated with test results
- [ ] README includes testing section

---

## Success Criteria

### Minimum Passing Threshold

- Unit tests: 6/6 files written and passing
- Integration tests: 4/4 files written and passing
- E2E tests: 1/1 file (optional, requires LM Studio)
- Total code coverage: 75%+ of non-test Go code

### Quality Standards

- All tests have clear test names describing behavior
- Tests cover both happy path and edge cases
- Each test runs in <30 seconds independently
- Tests don't depend on network (unless integration)
