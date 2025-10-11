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

Try the example program:
```bash
./blang -o hello.s hello.b
```
