# Testing Guide

## Quick Start

```bash
make test

# Run specific test file
go test -run TestExpressions
```

## Test Structure

**186 tests** across **11 organized files** with **78.5% coverage**:

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
| `cli_test.go` | **CLI interface** | **~33** |

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

## CLI Testing

The CLI test suite (`cli_test.go`) provides comprehensive testing of the command-line interface:

### Test Categories

| Test Function | Purpose | Test Count |
|---------------|---------|------------|
| `TestCLIBasicOptions` | Help, version, error handling | 5 tests |
| `TestCLIOutputFormats` | Different output formats | 5 tests |
| `TestCLIOptimizationFlags` | Optimization levels | 5 tests |
| `TestCLIDebugAndVerbose` | Debug and verbose options | 3 tests |
| `TestCLIWarningFlags` | Warning options | 3 tests |
| `TestCLIPathFlags` | Include/library paths | 4 tests |
| `TestCLIStandardFlag` | Language standard | 2 tests |
| `TestCLISaveTemps` | Temporary file handling | 2 tests |
| `TestCLIExecutableGeneration` | End-to-end compilation | 1 test |
| `TestCLICombinedFlags` | Complex flag combinations | 3 tests |

### Running CLI Tests

```bash
# Run all CLI tests
go test -v -run TestCLI

# Run specific CLI test categories
go test -v -run TestCLIBasicOptions
go test -v -run TestCLIOutputFormats
go test -v -run TestCLIOptimizationFlags

# Run CLI tests with coverage
go test -v -run TestCLI -cover
```

### CLI Test Features

- **Comprehensive flag testing**: Every CLI option is tested
- **Error condition testing**: Invalid inputs and error handling
- **File system testing**: Output file generation and cleanup
- **Executable verification**: End-to-end testing with real execution
- **Integration testing**: Works with actual `libb.o` runtime library

## Coverage

Current coverage: **78.5%**

See [Test_Coverage_Plan.md](Test_Coverage_Plan.md) for improvement roadmap.
