# B Language Compiler

A modern B programming language compiler written in Go with LLVM IR backend.

The [B programming language](https://en.wikipedia.org/wiki/B_(programming_language)) was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 as the predecessor to C.

**Status:** ✅ **Feature-Complete** • 145 tests passing • 76.6% coverage

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

**🎉 100% B Language Feature Completeness Achieved!**

**Core Language:**
- ✅ Variables (`auto`, `extrn`, global scalars)
- ✅ Arrays (local and global with B semantics)
- ✅ Pointers (address-of, dereference, arithmetic)
- ✅ Functions (definitions, parameters, recursion)

**All Operators (15 precedence levels):**
- ✅ Arithmetic: `+`, `-`, `*`, `/`, `%`
- ✅ Comparison: `<`, `<=`, `>`, `>=`, `==`, `!=`
- ✅ Bitwise: `&`, `|`, `<<`, `>>`
- ✅ Logical: `!`
- ✅ Unary: `-`, `&`, `*`, `++`, `--` (prefix & postfix)
- ✅ Assignment: `=`
- ✅ **Compound assignments:** `=+`, `=-`, `=*`, `=/`, `=%`, `=<<`, `=>>`, `=&`, `=|`, `=<`, `=<=`, `=>`, `=>=`, `=!=`, `===` (all 15 operators)
- ✅ **Ternary conditional:** `? :` (with nested support)
- ✅ Array indexing: `[]`
- ✅ Function calls: `()` (including indirect calls via pointers)

**Control Flow:**
- ✅ `if`/`else` statements (including nested)
- ✅ `while` loops (with unique labels for nesting)
- ✅ `switch`/`case` statements
- ✅ `goto` and labels
- ✅ `return` statements

**Advanced Features:**
- ✅ Scalar with multiple initialization values (`c -345, 'foo', "bar";`)
- ✅ Character constants in array sizes (`auto buf['x'];`)
- ✅ Reverse auto allocation order
- ✅ **Indirect function calls** via function pointer variables
- ✅ Forward references with `extrn`
- ✅ Multi-character constants
- ✅ All escape sequences
- ✅ **Optimized large array generation** (95-99% .ll file size reduction)

**Backend & Runtime:**
- ✅ LLVM IR code generation for portability
- ✅ Comprehensive runtime library (`printf`, `write`, `exit`, etc.)
- ✅ Automatic external function declaration

See **[LLVM_Backend.md](doc/LLVM_Backend.md)** for detailed feature documentation.

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

  **145 tests passing** (100% pass rate):

See **[Testing.md](doc/Testing.md)** for detailed testing guide and **[Test_Coverage_Plan.md](doc/Test_Coverage_Plan.md)** for coverage improvement roadmap (76.6% → 85%+ target).

## Documentation

- **[Testing Guide](doc/Testing.md)** - How to compile and run B programs, test examples
- **[LLVM Backend](doc/LLVM_Backend.md)** - LLVM IR code generation, all features documented
- **[Runtime Library](libb/README.md)** - Complete `libb.c` function reference
- **[Test Coverage Plan](doc/Test_Coverage_Plan.md)** - Detailed plan to improve coverage 76.6% → 85%+
- **[Development Journal](doc/Journal.md)** - Complete project history (C prototype → feature-complete Go compiler)
- **[TODO List](doc/TODO.md)** - All core features complete! Optional enhancements only

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

All core B language features are complete! Contributions welcome for:
- Improving test coverage (see **[Test_Coverage_Plan.md](doc/Test_Coverage_Plan.md)**)
- Code quality improvements
- Performance optimizations
- Additional platforms
- Documentation enhancements

Check **[TODO.md](doc/TODO.md)** for optional enhancement ideas.

## References

- [B Language Reference](https://github.com/sergev/blang/raw/refs/heads/main/doc/bref.pdf) - Original manual by S.C.Johnson
- [B Tutorial](https://github.com/sergev/blang/raw/refs/heads/main/doc/btut.pdf) - A tutorial introduction by B.W.Kernighan
- [Users' Reference to B](https://github.com/sergev/blang/raw/refs/heads/main/doc/kbman.pdf) - Ken Thompson's guide
- [LLVM Documentation](https://llvm.org/docs/LangRef.html) - LLVM IR reference
- [BCause](https://github.com/Spydr06/BCause) - C-based B compiler (inspiration)

## License

See LICENSE file.

---

**Implementation:** Go 1.21+ | **Backend:** LLVM IR | **Platforms:** macOS, Linux (x86_64)
