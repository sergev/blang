# Test Coverage Plan

## Current Status

**Coverage**: 80.0%  
**Tests**: 153 tests  
**Target**: 85%+

## Test Organization

**10 focused test files** (2,437 lines total):

| File | Lines | Purpose |
|------|-------|---------|
| `compiler_test.go` | 233 | Basic compilation |
| `expressions_test.go` | 792 | Operators & expressions |
| `functions_test.go` | 297 | Function calls |
| `operators_test.go` | 238 | Compound operators |
| `examples_test.go` | 177 | Integration tests |
| `lexer_test.go` | 204 | Token parsing |
| `control_flow_test.go` | 113 | Loops & control |
| `runtime_test.go` | 141 | Runtime library |
| `globals_test.go` | 138 | Global variables |
| `strings_test.go` | 104 | String handling |

## Coverage Gaps

### Medium Priority (50-70% Coverage)
- Error handling paths
- Edge cases in parsing
- Complex expression combinations

## Improvement Plan

1. **Add error path tests** - Test invalid input handling
2. **Add edge case tests** - Boundary conditions, unusual inputs
3. **Add integration tests** - End-to-end scenarios
4. **Add performance tests** - Large programs, stress testing

## Target: 85%+ Coverage

Focus on untested code paths and error handling scenarios.