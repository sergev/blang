The goal of the project is to build a compiler for the B programming language.
The compiler is written in Go, and emits IR assembly for LLVM.

The [B programming language](https://en.wikipedia.org/wiki/B_(programming_language))
was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 and was later replaced by C.

Implementation is loosly based on [BCause compiler](https://github.com/Spydr06/BCause) by Spydr06.

## 📚 Documentation

- **[LLVM Backend Details](doc/LLVM_Backend.md)** - LLVM IR code generation and linking
- **[Testing Guide](doc/Testing.md)** - Running tests and examples
- **[TODO List](doc/TODO.md)** - Pending features and roadmap
- **[Development Journal](doc/Journal.md)** - Complete development history
- **[Runtime Library](libb/README.md)** - B standard library functions (`printf`, `write`, etc.)

## Building

To build the compiler:
```bash
make
```

Or using Go directly:
```bash
go build -o blang
```

## Usage

To compile a B source file (`.b`):
```bash
./blang -o output.s input.b
```

The compiler generates IR assembly code that can be assembled and linked using standard LLVM tools.

To get help:
```bash
./blang --help
```

## Features

The compiler supports the full B language including:
- Global and local variables
- Arrays (vectors)
- Functions with arguments
- Control flow: if/else, while, switch/case, goto
- Expressions with proper operator precedence
- Multi-character constants
- String literals
- Comments (/* ... */)

## Testing

> **Note**: Unit tests are temporarily disabled during LLVM backend migration. Tests will be re-enabled incrementally as features are fully implemented.

The project includes comprehensive unit tests covering:
- **Compilation tests**: Full end-to-end compilation of various B programs
- **Error handling tests**: 8 different error scenarios (undefined variables, unclosed comments, etc.)
- **Lexer tests**: Tokenization, identifiers, numbers, strings, characters, comments

Run all tests:
```bash
go test
```

Run tests with verbose output:
```bash
go test -v
```

Run tests with coverage:
```bash
go test -cover
```

Run benchmarks:
```bash
go test -bench=. -benchmem
```

### Test Programs

The `testdata/` directory contains various B programs for testing:
- `hello.b` - Simple hello world
- `arithmetic.b` - Arithmetic operations and function calls
- `globals.b` - Global variables and arrays
- `conditionals.b` - if/else statements
- `loops.b` - while loops and factorial
- `strings.b` - String literals
- `operators.b` - Bitwise, comparison, and unary operators

Try the example programs:
```bash
./blang -o hello.ll testdata/hello.b
./blang -o arithmetic.ll testdata/arithmetic.b
```

## Current Status

✅ **LLVM IR Backend** - Complete
✅ **Expression Parser with Full Operator Precedence** - Complete (15 levels)
✅ **Operators** - Fully Implemented
  - Arithmetic: `+`, `-`, `*`, `/`, `%`
  - Comparison: `<`, `<=`, `>`, `>=`, `==`, `!=`
  - Bitwise: `&`, `|`, `<<`, `>>`
  - Logical: `!`
  - Unary: `-`, `&` (address-of), `*` (dereference)
  - Increment/Decrement: `++`, `--` (prefix and postfix)
  - Assignment: `=`
  - Ternary: `? :`
✅ **Control Flow** - Complete
  - if/else statements
  - while loops
  - switch/case statements (LLVM switch instruction)
  - goto statements
  - Labels
  - Recursive function calls
✅ **Functions** - Complete
  - Declarations and definitions
  - Function calls with parameters
  - Return values
  - Automatic external function declaration
✅ **Variables** - Complete (local `auto`, global, `extrn`)
✅ **String Literals** - Complete
✅ **Arrays** - Complete
  - Local arrays with automatic initialization
  - Global arrays with constant expressions
  - Array indexing with `[]` operator (auto-scales)
  - Arrays as function parameters
✅ **Pointers** - Complete
  - Address-of operator (`&`)
  - Dereference operator (`*`)
  - Pointer indexing with `[]`
  - Stores through pointers
⏳ **Compound Assignment** - Pending (`=+`, `=-`, etc. - use `x = x + 5` instead)  
⏳ **Indirect Function Calls** - Pending (extrn function pointers - use direct calls for now)  
⏳ **Unit Tests** - Temporarily disabled during migration (to be re-enabled incrementally)

### Important Notes

**Function Calls:**
- ✅ `printf("hello")` - Undefined function, auto-declared as external (works!)
- ⏳ `extrn printf; printf("hello")` - Function pointer variable, indirect call (pending)
- 💡 **Recommendation**: Use direct function calls (no `extrn` for functions)

### Verified Working Programs:

All test programs compile to LLVM IR and execute correctly:

| Program | Description | Result |
|---------|-------------|--------|
| `hello.b` | printf with strings | ✅ Returns 0 |
| `arithmetic.b` | All arithmetic operators | ✅ Returns 50 |
| `conditionals.b` | if/else, max, abs functions | ✅ Returns 35 |
| `loops.b` | While loop, factorial(5) | ✅ Returns 120 |
| `arrays.b` | Array operations, sum function | ✅ Returns 150 |
| `pointers.b` | Pointer ops, &, *, indexing | ✅ Returns 30 |
| `globals.b` | Global variables & arrays | ✅ Returns 60 |
| `switch.b` | Switch/case statements | ✅ Returns 30 |
| `goto.b` | goto and labels | ✅ Returns 42 |

### Linking with Runtime Library

```bash
# First time: compile the B runtime library
clang -c -ffreestanding libb/libb.c -o libb.o

# Compile B program to LLVM IR
./blang -o program.ll program.b

# Link and create executable  
clang program.ll libb.o -o program

# Run
./program
```

**Important**: The B runtime library (`libb/libb.c`) provides standard B functions:
- **`write(c)`** - For multi-character constants like `write('Hi!*n')`
- **`printf(fmt, ...)`** - For string pointers like `printf("Hello %s", name)`
- See **[libb/README.md](libb/README.md)** for complete runtime library documentation

---

## Contributing

See **[TODO.md](doc/TODO.md)** for a list of pending features and suggested next steps.

## Additional Resources

- **[LLVM Backend Architecture](doc/LLVM_Backend.md)** - Technical details of LLVM IR generation
- **[Development Journal](doc/Journal.md)** - Complete history of the project from C prototype to production-ready Go compiler
