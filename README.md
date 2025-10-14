# B Language Compiler

A modern B programming language compiler written in Go with LLVM IR backend.

The [B programming language](https://en.wikipedia.org/wiki/B_(programming_language)) was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 as the predecessor to C.

**Status:** ✅ **Feature-Complete** • 153 tests passing • 80.0% coverage

## Quick Start

```bash
# Build everything
make

# Compile and run a B program
./blang -o hello.ll examples/helloworld.b
clang hello.ll libb.o -o hello
./hello
```

**Output:** `Hello, World!`

## Features

- **Complete B Language Support**: All B language features implemented
- **LLVM IR Backend**: Portable, optimized code generation
- **Comprehensive Testing**: 153 tests across 10 organized test files
- **Modern Go Implementation**: Clean, maintainable codebase

## Example Programs

- `examples/hello.b` - Hello world using write()
- `examples/fibonacci.b` - Fibonacci calculator
- `examples/fizzbuzz.b` - FizzBuzz 1-100
- `examples/e-2.b` - E-2 constant calculation

## Testing

```bash
make test
```

## Documentation

- [Testing Guide](doc/Testing.md) - How to run tests
- [TODO](doc/TODO.md) - Future enhancements
- [B Language Reference](https://github.com/sergev/blang/raw/refs/heads/main/doc/bref.pdf) - Original manual by S.C.Johnson
- [B Tutorial](https://github.com/sergev/blang/raw/refs/heads/main/doc/btut.pdf) - A tutorial introduction by B.W.Kernighan
- [Users' Reference to B](https://github.com/sergev/blang/raw/refs/heads/main/doc/kbman.pdf) - Ken Thompson's guide
- [LLVM Documentation](https://llvm.org/docs/LangRef.html) - LLVM IR reference
- [BCause](https://github.com/Spydr06/BCause) - C-based B compiler (inspiration)

## License

MIT License - see [LICENSE](LICENSE) file.
