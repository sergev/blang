# blang Command-Line Interface Guide

The `blang` compiler provides a comprehensive command-line interface with multiple output formats, optimization levels, and debugging options.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Output Formats](#output-formats)
- [Optimization Options](#optimization-options)
- [Debugging and Verbose Output](#debugging-and-verbose-output)
- [Library Options](#library-options)
- [Other Options](#other-options)
- [Examples](#examples)
- [Error Handling](#error-handling)

## Basic Usage

```bash
blang [options] file...
```

### Input Files

- **Required**: At least one input file
- **Allowed extensions**: `.b`, `.ll`, `.s`, `.o`, `.a`
- **Behavior**:
  - `.b` files are compiled to LLVM IR automatically and then linked
  - `.ll`, `.s`, `.o`, `.a` files are passed to the system linker (`clang`)
- **Multiple files**: Any number of input files is supported; mixing extensions is allowed

```bash
# Single file
blang hello.b -o hello

# Multiple files
blang main.b utils.b -o program
```

### No Arguments

When called without any arguments, `blang` displays a concise usage message:

```bash
blang
# Output: Usage: blang [options] file...
```

## Output Formats

### Default: Executable

```bash
blang hello.b                    # Creates 'hello'
blang hello.b -o custom_name     # Creates 'custom_name'
```

Generates a fully linked executable. **Default output naming** uses the basename of the first source file (without directory path or extension) when `-o` is not provided.

### Object Files (`-c`)

```bash
blang -c hello.b                 # Creates 'hello.o'
blang -c hello.b -o custom.o     # Creates 'custom.o'
```

Compiles to object file without linking. Useful for separate compilation units and libraries.

Notes:
- When `-o` is provided, exactly one input file is required.
- Without `-o`, one `.o` is produced per input file in the current directory.

### Assembly (`-S`)

```bash
blang -S hello.b                 # Creates 'hello.s'
blang -S hello.b -o custom.s     # Creates 'custom.s'
```

Generates assembly code for understanding code generation and debugging.

Notes:
- Accepts `.b` and `.ll` inputs.
- When `-o` is provided, exactly one input file is required.
- Without `-o`, one `.s` is produced per input file in the current directory.

### LLVM IR (`--emit-llvm`)

```bash
blang --emit-llvm hello.b        # Creates 'hello.ll'
blang --emit-llvm hello.b -o custom.ll  # Creates 'custom.ll'
```

Generates LLVM Intermediate Representation for LLVM-based optimization and tool integration.

Notes:
- Accepts `.b` inputs only in this mode.
- When `-o` is provided, exactly one input file is required.
- Without `-o`, one `.ll` is produced per input file in the current directory.

## Optimization Options

### Optimization Levels

```bash
blang -O0 hello.b    # No optimization (default)
blang -O1 hello.b    # Basic optimization
blang -O2 hello.b    # Moderate optimization
blang -O3 hello.b    # Aggressive optimization
```

### Optimization Behavior

- **-O0**: Fastest compilation, no optimization
- **-O1**: Basic optimizations, faster compilation
- **-O2**: Moderate optimizations, balanced compilation time
- **-O3**: Aggressive optimizations, slower compilation

## Debugging and Verbose Output

### Debug Information (`-g`)

```bash
blang -g hello.b -o hello
```

Includes debug information in the generated executable for source-level debugging.

### Verbose Output (`-v`)

```bash
blang -v hello.b -o hello
```

Shows detailed compilation steps:

```
blang: compiling 1 file(s)
blang: processing hello.b
blang: generated hello.tmp.ll
blang: running clang hello.tmp.ll -Lruntime -lb -o hello
```

## Library Options

### Library Directories (`-L`)

```bash
blang -L /usr/lib -L /usr/local/lib hello.b -o hello
```

Adds directories to the library search path. **Can be repeated** for multiple directories.

### Link Libraries (`-l`)

```bash
blang -l pthread -l math hello.b -o hello
```

Links with specified libraries. **Can be repeated** for multiple libraries.

### Multiple Libraries

```bash
blang -L /usr/lib -L /usr/local/lib -l pthread -l math hello.b -o hello
```

## Other Options

### Save Temporary Files (`--save-temps`)

```bash
blang --save-temps hello.b -o hello
```

Preserves intermediate files (like `.tmp.ll` files) for debugging.

Temporary IR naming:
- Single `.b` with `-o <output>`: `<output>.tmp.ll`
- Multiple inputs: `<basename>.tmp.<idx>.ll` per `.b` input

### Help and Version

```bash
blang --help      # Show help information
blang -h          # Short form of --help
blang --version   # Show version information
blang -V          # Short form of --version
```

## Examples

### Development Workflow

```bash
# Debug build with verbose output
blang -g -O0 -v hello.b -o debug_hello

# Release build with optimization
blang -O2 hello.b -o release_hello

# Object file for library
blang -c -O2 hello.b -o hello.o
```

### Flexible Option Ordering

Options can be placed after arguments:

```bash
blang hello.b -o output -O2 -v -g
blang -v hello.b -o optimized -O3
```

### Multiple Output Formats

```bash
# Generate all formats
blang --emit-llvm hello.b        # hello.ll
blang -S hello.b                 # hello.s
blang -c hello.b                 # hello.o
blang hello.b                    # hello (executable)
```

### Library Integration

```bash
# With multiple library paths and libraries
blang -L /usr/lib -L /usr/local/lib -l pthread -l math hello.b -o hello
```

## Error Handling

### Common Error Types

1. **No input files**:
   ```
   blang: error: no input files
   compilation terminated.
   ```

2. **Invalid file extension**:
   ```
   blang: error: unsupported input file extension for 'test.txt'; allowed: .b, .ll, .s, .o, .a
   ```

3. **File not found**:
   ```
   blang: error: cannot access file 'missing.b': no such file or directory
   ```

4. **Invalid optimization level**:
   ```
   blang: error: invalid optimization level: 5
   ```

5. **Missing runtime library**:
   ```
   blang: error: runtime/libb.a not found (or clang cannot find -lb); run 'make' in runtime/
   ```

### Error Message Format

Error messages follow this format:
```
blang: error: <description>
compilation terminated.
```

## Integration with Build Systems

### Makefile

```makefile
CC = blang
CFLAGS = -O2 -g
TARGET = program
SOURCES = main.b utils.b

$(TARGET): $(SOURCES)
	$(CC) $(CFLAGS) -o $(TARGET) $(SOURCES)

clean:
	rm -f $(TARGET) *.o *.ll *.s
```

### CMake

```cmake
set(CMAKE_C_COMPILER blang)
set(CMAKE_C_FLAGS "-O2 -g")
add_executable(hello hello.b)
```

## Best Practices

1. **Use meaningful output names**: Always specify `-o` for clarity
2. **Optimize for release**: Use `-O2` or `-O3` for production builds
3. **Include debug info**: Use `-g` for debugging builds
4. **Use verbose mode**: Use `-v` to understand compilation steps
5. **Leverage default naming**: Let `blang` use basename-based defaults when appropriate
6. **Flexible ordering**: Place options where they're most readable

## Troubleshooting

### Common Issues

1. **"runtime/libb.a not found" or "cannot find -lb"**: Run `make` in `runtime/` to build the runtime library
2. **"clang not found"**: Ensure clang is installed and in PATH
3. **Permission denied**: Ensure output directory is writable
4. **Linker errors**: Check that required libraries are available

### Debugging Compilation

1. Use `-v` to see compilation steps
2. Use `--save-temps` to examine intermediate files
3. Use `--emit-llvm` to see generated LLVM IR
4. Check that input files are valid B source code

## Compatibility

The `blang` CLI is designed to be compatible with:
- Standard Unix/Linux build tools
- Modern build systems (CMake, Meson, etc.)
- IDE integration
- Continuous integration systems

This makes it easy to integrate `blang` into existing development workflows.