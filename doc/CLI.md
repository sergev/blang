# blang Command-Line Interface Guide

The `blang` compiler provides a comprehensive command-line interface similar to `clang`, supporting multiple output formats, optimization levels, debugging options, and more.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Output Formats](#output-formats)
- [Optimization Options](#optimization-options)
- [Debugging and Verbose Output](#debugging-and-verbose-output)
- [Warning and Error Handling](#warning-and-error-handling)
- [Path and Library Options](#path-and-library-options)
- [Other Options](#other-options)
- [Examples](#examples)
- [Error Messages](#error-messages)
- [Integration with Build Systems](#integration-with-build-systems)

## Basic Usage

```bash
blang [options] file...
```

### Input Files

- **Required**: At least one `.b` source file
- **Extension**: Input files must have `.b` extension
- **Multiple files**: Multiple source files can be compiled together

```bash
# Single file
blang -o hello hello.b

# Multiple files
blang -o program main.b utils.b
```

## Output Formats

### Default: Executable

```bash
blang -o hello hello.b
```

Generates a fully linked executable that can be run directly.

### Object Files (`-c`)

```bash
blang -c -o hello.o hello.b
```

Compiles to object file without linking. Useful for:
- Separate compilation units
- Creating libraries
- Manual linking control

### Assembly (`-S`)

```bash
blang -S -o hello.s hello.b
```

Generates assembly code. Useful for:
- Understanding code generation
- Manual assembly optimization
- Debugging compilation issues

### LLVM IR (`-emit-llvm`)

```bash
blang -emit-llvm -o hello.ll hello.b
```

Generates LLVM Intermediate Representation. Useful for:
- LLVM-based optimization
- Integration with other LLVM tools
- Understanding the compilation process

### Preprocessed (`-E`)

```bash
blang -E -o hello.i hello.b
```

Performs preprocessing only. Currently outputs the source unchanged (B language has minimal preprocessing).

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

The optimization flags are passed through to the LLVM backend and linker.

## Debugging and Verbose Output

### Debug Information (`-g`)

```bash
blang -g -o hello hello.b
```

Includes debug information in the generated executable, enabling:
- Source-level debugging with `gdb` or `lldb`
- Symbol information
- Line number mapping

### Verbose Output (`-v`)

```bash
blang -v -o hello hello.b
```

Shows detailed compilation steps:

```
blang: compiling 1 file(s)
blang: processing hello.b
blang: generated hello.tmp.ll
blang: running /usr/bin/clang hello.tmp.ll libb.o -o hello
blang: generated hello
```

## Warning and Error Handling

### Enable Warnings (`-Wall`)

```bash
blang -Wall -o hello hello.b
```

Enables comprehensive warning messages for:
- Potential issues in the code
- Non-standard language usage
- Optimization opportunities

### Treat Warnings as Errors (`-Werror`)

```bash
blang -Werror -o hello hello.b
```

Causes compilation to fail if any warnings are generated.

### Combined Warning Options

```bash
blang -Wall -Werror -o hello hello.b
```

## Path and Library Options

### Include Directories (`-I`)

```bash
blang -I /path/to/headers -o hello hello.b
```

Adds directories to the include search path (currently minimal support in B language).

### Library Directories (`-L`)

```bash
blang -L /path/to/libs -o hello hello.b
```

Adds directories to the library search path for linking.

### Link Libraries (`-l`)

```bash
blang -l math -o hello hello.b
```

Links with specified libraries (passed to the linker).

### Multiple Paths

```bash
blang -I /usr/include,/usr/local/include -L /usr/lib,/usr/local/lib -o hello hello.b
```

Comma-separated paths are supported for convenience.

## Other Options

### Language Standard (`-std`)

```bash
blang -std b -o hello hello.b
```

Specifies the language standard (currently only `b` is supported).

### Save Temporary Files (`-save-temps`)

```bash
blang -save-temps -o hello hello.b
```

Preserves intermediate files (like `.tmp.ll` files) for debugging.

### Output File (`-o`)

```bash
blang -o custom_name hello.b
```

Specifies the output filename. Default names:
- **Executable**: `a.out`
- **Object file**: `filename.o`
- **Assembly**: `filename.s`
- **LLVM IR**: `filename.ll`
- **Preprocessed**: `filename.i`

### Help and Version

```bash
blang -help      # Show help information
blang -version   # Show version information
```

## Examples

### Development Workflow

```bash
# Debug build
blang -g -O0 -Wall -v -o debug_hello hello.b

# Release build
blang -O2 -o release_hello hello.b

# Object file for library
blang -c -O2 -o utils.o utils.b
```

### Integration with Make

```makefile
CC = blang
CFLAGS = -Wall -O2
TARGET = program
SOURCES = main.b utils.b

$(TARGET): $(SOURCES)
	$(CC) $(CFLAGS) -o $(TARGET) $(SOURCES)

clean:
	rm -f $(TARGET) *.o *.ll *.s
```

### Multi-stage Compilation

```bash
# Step 1: Generate LLVM IR
blang -emit-llvm -o hello.ll hello.b

# Step 2: Generate object file
llc -filetype=obj -o hello.o hello.ll

# Step 3: Link manually
clang hello.o libb.o -o hello
```

## Error Messages

### Common Error Types

1. **File not found**:
   ```
   blang: error: cannot access file 'missing.b': no such file or directory
   ```

2. **Invalid file extension**:
   ```
   blang: error: input file 'test.txt' does not have .b extension
   ```

3. **Syntax errors**:
   ```
   blang: error: unclosed char literal
   ```

4. **Missing runtime library**:
   ```
   blang: error: libb.o not found, run 'make' first
   ```

### Error Message Format

Error messages follow this format:
```
blang: error: <description>
compilation terminated.
```

## Integration with Build Systems

### CMake

```cmake
set(CMAKE_C_COMPILER blang)
set(CMAKE_C_FLAGS "-Wall -O2")
add_executable(hello hello.b)
```

### Meson

```meson
project('hello', 'c')
executable('hello', 'hello.b', c_args: ['-Wall', '-O2'])
```

### Autotools

```makefile
CC = blang
CFLAGS = -Wall -O2
hello: hello.b
	$(CC) $(CFLAGS) -o $@ $<
```

## Best Practices

1. **Use meaningful output names**: Always specify `-o` for clarity
2. **Enable warnings**: Use `-Wall` during development
3. **Optimize for release**: Use `-O2` or `-O3` for production builds
4. **Include debug info**: Use `-g` for debugging builds
5. **Clean temporary files**: Don't use `-save-temps` unless needed
6. **Use verbose mode**: Use `-v` to understand compilation steps

## Troubleshooting

### Common Issues

1. **"libb.o not found"**: Run `make` to build the runtime library
2. **"llc not found"**: LLVM tools are not required for basic compilation
3. **Permission denied**: Ensure output directory is writable
4. **Linker errors**: Check that required libraries are available

### Debugging Compilation

1. Use `-v` to see compilation steps
2. Use `-save-temps` to examine intermediate files
3. Use `-emit-llvm` to see generated LLVM IR
4. Check that input files are valid B source code

## Compatibility

The `blang` CLI is designed to be compatible with:
- Standard Unix/Linux build tools
- Modern build systems (CMake, Meson, etc.)
- IDE integration
- Continuous integration systems

This makes it easy to integrate `blang` into existing development workflows.
