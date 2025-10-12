The goal of the project is to build a compiler for the B programming language.
The compiler is written in Go, and emits IR assembly for LLVM.

The [B programming language](https://en.wikipedia.org/wiki/B_(programming_language))
was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 and was later replaced by C.

Implementation is loosly based on [BCause compiler](https://github.com/Spydr06/BCause) by Spydr06.

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
- **List tests**: Dynamic list data structure operations

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
✅ **Basic Compilation** - Working (functions, variables, return statements)  
✅ **String Literals** - Working  
✅ **Function Calls** - Working  
⚠️ **Full Expression Support** - In Progress (assignments working, operators pending)  
⚠️ **Control Flow** - In Progress (if/while implemented, switch/goto pending)  
⚠️ **Arrays** - Partial (global arrays declared, indexing pending)  
⏳ **Unit Tests** - Temporarily disabled during migration
