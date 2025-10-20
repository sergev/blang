# B Language Runtime Library (libb)

**Platform:** macOS, Linux (x86_64, ARM64, RISC-V)
**Language:** C (freestanding)

## Overview

The B runtime library provides essential I/O operations for B language programs. It's freestanding (no C standard library dependency) and uses direct system calls, making it lightweight and portable across multiple architectures.

## Architecture Support

**Supported Architectures:**
- **x86_64:** `syscall` instruction (macOS: +0x2000000 offset)
- **ARM64:** `svc` instruction (Linux: x8 register, macOS: x16 register)
- **RISC-V:** `ecall` instruction (standard calling convention)

**Function Aliasing:** All functions use platform-specific aliases (`b.name` on Linux, `_b.name` on macOS) to avoid conflicts with system libraries.

## Functions Reference

### I/O Functions

| Function | Description | Example |
|----------|-------------|---------|
| `write(c)` | Write multi-character constant (big-endian packed) | `write('Hello')` |
| `writeb(c)` | Write single byte | `writeb('A')` |
| `printf(fmt, ...)` | Formatted output (%d, %o, %c, %s, %%) | `printf("Value: %d*n", 42)` |
| `read()` | Read character from stdin (ASCII only) | `c = read()` |
| `nread(fd, buf, n)` | Read n bytes from file descriptor | `nread(0, buffer, 100)` |
| `nwrite(fd, buf, n)` | Write n bytes to file descriptor | `nwrite(1, buffer, n)` |

### String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `char(s, i)` | Get i-th character from string | `c = char("Hello", 0)` |
| `lchar(s, i, c)` | Set i-th character in string | `lchar(s, 0, 'h')` |

### System Functions

| Function | Description |
|----------|-------------|
| `exit()` | Terminate process (exit code 0) |
| `flush()` | No-op (all I/O is unbuffered) |
| `start()` | Program entry point (Linux only) |

### Helper Functions

| Function | Description |
|----------|-------------|
| `printd(n)` | Print decimal number (recursive) |
| `printo(n)` | Print octal number (bit-shifting) |

## Global Variables

**`b_fout`** - Output file descriptor offset:
- `b_fout = 0` → stdout (fd 1)
- `b_fout = 1` → stderr (fd 2)

## Compilation

```bash
make libb.a
```

## Linking

```bash
# Compile B program
./blang program.b -o program.ll

# Link with runtime
clang program.ll -L. -lb -o program
```

## Variadic ABI

All functions use variadic declarations for consistent calling convention:
```c
word_t b_char(word_t string, ...) ALIAS("char");
void b_writeb(word_t c, ...) ALIAS("writeb");
```

## Platform Details

- **macOS:** Requires `-ffreestanding` flag
- **Linux:** Includes custom `_start` entry point
- **All platforms:** Automatic architecture detection via compiler macros

## Recent Changes

- ✅ **Multi-Architecture:** ARM64 and RISC-V support
- ✅ **Variadic ABI:** Uniform function interface across all architectures
- ✅ **Function Aliasing:** Platform-specific name resolution
- ✅ **Build System:** Cross-platform compilation support

## Technical Specs

- **Size:** ~2.6 KB (`libb.a`)
- **Dependencies:** None (freestanding)
- **System Calls:** read, write, exit
- **All I/O:** Unbuffered (direct system calls)
