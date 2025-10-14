# B Language Compiler

A modern B programming language compiler written in Go with LLVM IR backend and clang-like command-line interface.

The [B programming language](https://en.wikipedia.org/wiki/B_(programming_language)) was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 as the predecessor to C.

**Status:** ✅ **Feature-Complete** • 195 tests passing • 71.1% coverage

## Quick Start

```bash
# Build everything
make

# Compile and run a B program (automatic linking)
./blang -o hello examples/helloworld.b
./hello
```

**Output:** `Hello, World!`

## Features

- **Complete B Language Support**: All B language features implemented
- **Clang-like CLI Interface**: Professional command-line options and workflow
- **Multiple Output Formats**: Executable, object files, assembly, LLVM IR
- **Automatic Linking**: Seamless integration with runtime library
- **LLVM IR Backend**: Portable, optimized code generation
- **Comprehensive Testing**: 195 tests across 11 organized test files
- **Modern Go Implementation**: Clean, maintainable codebase

## Command-Line Interface

The `blang` compiler provides a clang-like command-line interface with comprehensive options:

### Basic Usage

```bash
# Generate executable (default)
blang -o hello hello.b

# Generate LLVM IR
blang -emit-llvm -o hello.ll hello.b

# Generate object file
blang -c -o hello.o hello.b

# Generate assembly
blang -S -o hello.s hello.b

# Preprocess only
blang -E -o hello.i hello.b
```

### Compiler Options

| Option | Description |
|--------|-------------|
| `-o <file>` | Place output into `<file>` |
| `-c` | Compile and assemble, but do not link |
| `-S` | Compile only; do not assemble or link |
| `-E` | Preprocess only; do not compile, assemble or link |
| `-emit-llvm` | Emit LLVM IR instead of executable |

### Optimization and Debugging

| Option | Description |
|--------|-------------|
| `-O0`, `-O1`, `-O2`, `-O3` | Optimization levels |
| `-g` | Generate debug information |
| `-v` | Verbose output |

### Warnings and Diagnostics

| Option | Description |
|--------|-------------|
| `-Wall` | Enable all warnings |
| `-Werror` | Treat warnings as errors |

### Paths and Libraries

| Option | Description |
|--------|-------------|
| `-I<dir>` | Add directory to include search path |
| `-L<dir>` | Add directory to library search path |
| `-l<lib>` | Link with library |

### Other Options

| Option | Description |
|--------|-------------|
| `-std=<standard>` | Language standard (default: `b`) |
| `-save-temps` | Do not delete intermediate files |
| `-help` | Display help information |
| `-version` | Display version information |

### Examples

```bash
# Optimized executable with debug info
blang -O2 -g -o optimized hello.b

# Verbose compilation with warnings
blang -v -Wall -o hello hello.b

# Generate object file for linking
blang -c -o hello.o hello.b

# All flags combined
blang -v -O3 -g -Wall -save-temps -o hello hello.b
```

## Example Programs

- `examples/hello.b` - Hello world using write()
- `examples/fibonacci.b` - Fibonacci calculator
- `examples/fizzbuzz.b` - FizzBuzz 1-100
- `examples/e-2.b` - E-2 constant calculation

## Testing

```bash
# Run all tests
make test

# Run CLI-specific tests
go test -v -run TestCLI

# Run specific test categories
go test -v -run TestCLIBasicOptions
go test -v -run TestCLIOutputFormats
```

## Documentation

- [CLI Usage Guide](doc/CLI.md) - Comprehensive command-line interface guide
- [Testing Guide](doc/Testing.md) - How to run tests and CLI test suite
- [TODO](doc/TODO.md) - Future enhancements
- [B Language Reference](https://github.com/sergev/blang/raw/refs/heads/main/doc/bref.pdf) - Original manual by S.C.Johnson
- [B Tutorial](https://github.com/sergev/blang/raw/refs/heads/main/doc/btut.pdf) - A tutorial introduction by B.W.Kernighan
- [Users' Reference to B](https://github.com/sergev/blang/raw/refs/heads/main/doc/kbman.pdf) - Ken Thompson's guide
- [LLVM Documentation](https://llvm.org/docs/LangRef.html) - LLVM IR reference
- [BCause](https://github.com/Spydr06/BCause) - C-based B compiler (inspiration)

## License

MIT License - see [LICENSE](LICENSE) file.
