# B Language Compiler

A modern B programming language compiler written in Go with LLVM IR backend and clang-like command-line interface.

The [B programming language](https://en.wikipedia.org/wiki/B_(programming_language)) was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 as the predecessor to C.

**Status:** ✅ **Feature-Complete** • 225 tests passing • 75.1% coverage

## Quick Start

```bash
# Build everything
make

# Compile and run a B program (automatic linking)
./blang examples/hello.b
./hello
```

**Output:** `Hello, World!`

## Installation

### Prerequisites
- Go (>=1.21)
- clang/LLVM toolchain
- make

### Install (user-local)
```bash
make
make install               # installs to $HOME/.local
# ensure PATH contains $HOME/.local/bin
export PATH="$HOME/.local/bin:$PATH"   # add to your shell profile
blang -V
```

### Uninstall
```bash
make uninstall
```

### Custom prefix
```bash
make install DESTDIR=/custom/prefix
```

## Features

- **Complete B Language Support**: All B language features implemented
- **Cross-platform**: Runs natively on Linux (x86_64, ARM64, RISC-V) and macOS (x86_64, ARM64)
- **Clang-like CLI Interface**: Professional command-line options and workflow
- **Multiple Output Formats**: Executable, object files, assembly, LLVM IR
- **LLVM IR Backend**: Portable, optimized code generation
- **Comprehensive Testing**: 225 tests across 8 organized test files
- **Modern Go Implementation**: Clean, maintainable codebase

## Command-Line Interface

The `blang` compiler provides a clang-like command-line interface with comprehensive options:

### Basic Usage

```bash
# Generate executable
blang hello.b -o hello

# Generate LLVM IR
blang -emit-llvm hello.b -o hello.ll

# Generate object file
blang -c hello.b -o hello.o

# Generate assembly
blang -S hello.b -o hello.s
```

### Compiler Options

| Option | Description |
|--------|-------------|
| `-o <file>` | Place output into `<file>` |
| `-c` | Compile and assemble, but do not link |
| `-S` | Compile only; do not assemble or link |
| `-emit-llvm` | Emit LLVM IR instead of executable |

### Optimization and Debugging

| Option | Description |
|--------|-------------|
| `-O0`, `-O1`, `-O2`, `-O3` | Optimization levels |
| `-g` | Generate debug information |
| `-v` | Verbose output |

### Paths and Libraries

| Option | Description |
|--------|-------------|
| `-L <dir>` | Add directory to library search path (can be repeated) |
| `-l <lib>` | Link with library (can be repeated) |

### Other Options

| Option | Description |
|--------|-------------|
| `-save-temps` | Do not delete intermediate files |
| `-h`, `--help` | Display help information |
| `-V`, `--version` | Display version information |

### Examples

```bash
# Optimized executable with debug info
blang -O2 -g hello.b -o optimized

# Verbose compilation
blang -v hello.b

# Multiple library directories and libraries
blang hello.b -L/usr/lib -L/usr/local/lib -lpthread -lmath

# Options after arguments (flexible ordering)
blang hello.b -O2 -v

# All flags combined
blang -v -O3 -g -save-temps hello.b
```

## Example Programs

- `examples/hello.b` - Hello world using write()
- `examples/fibonacci.b` - Fibonacci calculator
- `examples/fizzbuzz.b` - FizzBuzz 1-100
- `examples/e-2.b` - E-2 constant calculation
- `examples/b.b` - B compiler for PDP-7

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
- [B Language Reference](https://github.com/sergev/blang/raw/refs/heads/main/doc/bref.pdf) - Original manual by S.C.Johnson
- [B Tutorial](https://github.com/sergev/blang/raw/refs/heads/main/doc/btut.pdf) - A tutorial introduction by B.W.Kernighan
- [Users' Reference to B](https://github.com/sergev/blang/raw/refs/heads/main/doc/kbman.pdf) - Ken Thompson's guide
- [LLVM Documentation](https://llvm.org/docs/LangRef.html) - LLVM IR reference
- [BCause](https://github.com/Spydr06/BCause) - C-based B compiler (inspiration)

## License

MIT License - see [LICENSE](LICENSE) file.
