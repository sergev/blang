# Testing Guide

## Quick Start

```bash
make test

# Run specific test file
go test -run TestExpressions
```

## Test Structure

**157 tests** across **10 organized files** with **78.5% coverage**:

| File | Purpose | Tests |
|------|---------|-------|
| `compiler_test.go` | Basic compilation | ~15 |
| `examples_test.go` | Integration tests | ~12 |
| `expressions_test.go` | Operators & expressions | ~46 |
| `functions_test.go` | Function calls | ~18 |
| `operators_test.go` | Compound operators | ~20 |
| `globals_test.go` | Global variables | ~10 |
| `runtime_test.go` | Runtime library | ~8 |
| `control_flow_test.go` | Loops & control | ~8 |
| `strings_test.go` | String handling | ~5 |
| `lexer_test.go` | Token parsing | ~15 |

## Example Programs

- `examples/hello.b` - Hello world using write()
- `examples/helloworld.b` - Hello world using printf()
- `examples/fibonacci.b` - Fibonacci calculator
- `examples/fizzbuzz.b` - FizzBuzz 1-100
- `examples/e-2.b` - E-2 constant calculation

## Test Programs

- `testdata/arithmetic.b` - Arithmetic operations
- `testdata/globals.b` - Global variables and arrays
- `testdata/conditionals.b` - if/else statements
- `testdata/loops.b` - while loops and factorial
- `testdata/strings.b` - String literals
- `testdata/operators.b` - Bitwise and comparison operators
- `testdata/pointers.b` - Pointer operations
- `testdata/arrays.b` - Array indexing
- `testdata/switch.b` - Switch/case statements
- `testdata/goto.b` - Goto and labels
- `testdata/comprehensive_ptr.b` - Complex pointer operations

## Coverage

Current coverage: **78.5%**

See [Test_Coverage_Plan.md](Test_Coverage_Plan.md) for improvement roadmap.
