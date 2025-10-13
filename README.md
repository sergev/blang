# B Language Compiler

A modern B programming language compiler written in Go with LLVM IR backend.

The [B programming language](https://en.wikipedia.org/wiki/B_(programming_language)) was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 as the predecessor to C.

**Status:** ✅ Production-ready • 84 tests passing • 71.4% coverage

## Quick Start

```bash
# Build the compiler and runtime library
make

# Compile and run a B program
./blang -o hello.ll examples/helloworld.b
clang hello.ll libb.o -o hello
./hello
```

**Output:** `Hello, World!`

## Features

**Fully Implemented:**
- ✅ Variables, arrays, pointers, functions
- ✅ All operators with correct precedence
- ✅ Control flow: if/else, while, switch/case, goto
- ✅ String literals, multi-character constants
- ✅ LLVM IR backend for portability
- ✅ Comprehensive runtime library

**Pending:**
- ⏳ Compound assignments (`=+`, `=-`, etc.)
- ⏳ Ternary operator (`? :`)

See **[TODO.md](doc/TODO.md)** for complete feature roadmap.

## Usage

```bash
# Compile B program
./blang -o output.ll input.b

# Link and run
clang output.ll libb.o -o output
./output
```

Use `./blang --help` for more options.

## Examples

### Fibonacci Numbers
```b
n 10;

main() {
    extrn n;
    printf("Fibonacci %d = %d*n", n, fib(n));
}

fib(n) {
    auto a, b, c, i;
    b = 1;
    while (i < n) {
        c = a + b;
        a = b;
        b = c;
        i = i + 1;
    }
    return(a);
}
```

More examples in the `examples/` directory: `hello.b`, `helloworld.b`, `fibonacci.b`, `fizzbuzz.b`, `e-2.b`.

## Testing

```bash
go test -v          # Run all tests
go test -cover      # With coverage report
```

**84 tests passing** covering lexer, parser, code generation, and full integration tests.

See **[Testing.md](doc/Testing.md)** for detailed testing guide and test programs.

## Documentation

- **[Testing Guide](doc/Testing.md)** - How to compile and run B programs, test examples
- **[LLVM Backend](doc/LLVM_Backend.md)** - LLVM IR code generation details
- **[Runtime Library](libb/README.md)** - Complete `libb.c` function reference
- **[TODO List](doc/TODO.md)** - Pending features and roadmap
- **[Development Journal](doc/Journal.md)** - Complete project history

## Project Structure

```
blang/
├── *.go                   # Compiler source (lexer, parser, codegen)
├── libb/                  # B runtime library
│   ├── libb.c             # Runtime implementation
│   └── README.md          # Runtime documentation
├── doc/                   # Documentation
├── testdata/              # Test programs
└── examples/              # Example B programs
```

## Contributing

Contributions welcome! Check **[TODO.md](doc/TODO.md)** for:
- Feature priorities
- Implementation suggestions
- Time estimates

## References

- [B Language Reference](https://www.bell-labs.com/usr/dmr/www/kbman.html) - Original manual
- [B Tutorial](https://www.bell-labs.com/usr/dmr/www/btut.html) - Ken Thompson's guide
- [LLVM Documentation](https://llvm.org/docs/LangRef.html) - LLVM IR reference
- [BCause](https://github.com/Spydr06/BCause) - C-based B compiler (inspiration)

## License

See LICENSE file.

---

**Implementation:** Go 1.21+ | **Backend:** LLVM IR | **Platforms:** macOS, Linux (x86_64)
