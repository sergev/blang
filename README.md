The goal of the project is to build a compiler for the B programming language.
The compiler is written in Go, and emits assembly code for X86_64 architecture.

The B programming language was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 and was later replaced by C.

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

The compiler generates x86_64 assembly code that can be assembled and linked using standard tools.

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

The project includes comprehensive unit tests covering:
- **Compilation tests**: Full end-to-end compilation of various B programs
- **Lexer tests**: Tokenization, identifiers, numbers, strings, characters, comments
- **List tests**: Dynamic list data structure operations
- **Code coverage**: 56.3% of statements

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
./blang -o hello.s testdata/hello.b
./blang -o arithmetic.s testdata/arithmetic.b
```

## Directory Structure

Directory c-prototype/ contains sources of a similar compiler, but written in C.
The Go implementation in the root directory is a complete rewrite that retains all functionality.
See c-prototype/README.md for details about the B language.
