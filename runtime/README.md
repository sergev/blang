# B Language Runtime Library (libb)

**Version:** 1.0
**Platform:** macOS, Linux (x86_64)
**Language:** C (freestanding)

---

## Overview

The `libb.c` file provides the essential runtime library for B language programs. Since B programs compile to LLVM IR and then to native code, they need a minimal runtime library to perform I/O operations and other basic tasks that interact with the operating system.

This library is **freestanding** (does not depend on the C standard library) and uses direct system calls for all operations, making it lightweight and portable.

---

## Architecture

### System Call Interface

The library implements a custom system call wrapper that works on both Linux and macOS:

```c
static inline SYSCALL_TYPE syscall(SYSCALL_TYPE n, SYSCALL_TYPE a1,
                                   SYSCALL_TYPE a2, SYSCALL_TYPE a3)
```

**Platform-specific behavior:**
- **macOS:** System call numbers are OR'd with `0x2000000`
- **Linux:** System call numbers used directly
- **Implementation:** Inline assembly for `syscall` instruction

---

## Data Types

### B_TYPE
The fundamental data type representing B's single data type:
- Default: `intptr_t` (64-bit signed integer on x86_64)
- Configurable via `#define B_TYPE`
- Used for all B values: integers, pointers, characters, booleans

### B_FN Macro
Allows prefixing/postfixing B function names to avoid conflicts:
```c
#define B_FN(name) name          // Default: no prefix
#define B_FN(name) __b##name     // Example: prefix with "__b"
```

---

## Functions Reference

### Core I/O Functions

#### `write(c)` - Write Multi-character Constant
```c
void write(B_TYPE c)
```
Writes one or more characters to stdout. In B, multi-character constants are packed into a single word with the most significant byte first.

**Example:**
```b
write('H');           /* Writes 'H' */
write('Hello');       /* Writes 'Hello' (5 chars packed in one word) */
write('*n');          /* Writes newline */
```

**Implementation:** Recursively unpacks bytes from high to low order.

---

#### `writeb(c)` - Write Single Byte
```c
void writeb(B_TYPE c)
```
Writes a single byte to stdout. Used internally by `write()`.

**System Call:** `SYS_write` with file descriptor 1 (stdout)

---

#### `printf(fmt, ...)` - Formatted Output
```c
void printf(B_TYPE fmt, ...)
```
General formatting and printing function with format specifiers.

**Supported Format Specifiers:**
- `%d` - Decimal integer (signed)
- `%o` - Octal integer (unsigned)
- `%c` - Single character
- `%s` - String (null-terminated)
- `%%` - Literal '%' character

**Example:**
```b
printf("Value: %d*n", 42);           /* Value: 42 */
printf("Hex: %o*n", 15);             /* Hex: 17 (octal) */
printf("Char: %c*n", 'A');           /* Char: A */
printf("String: %s*n", "Hello");     /* String: Hello */
```

**Note:** Uses `*n` for newline (not `\n`) as B uses `*` for escape sequences.

**Implementation:**
- Uses C varargs (`va_list`)
- Calls helper functions `printd()` and `printo()` for number formatting
- Calls `char()` to extract string characters

---

### Helper Printing Functions

#### `printd(n)` - Print Decimal Number
```c
void printd(B_TYPE n)
```
Prints a signed decimal number. Used internally by `printf` for `%d` format.

**Features:**
- Handles negative numbers
- Recursive digit extraction
- No buffering (direct output)

---

#### `printo(n)` - Print Octal Number
```c
void printo(B_TYPE n)
```
Prints an unsigned octal number. Used internally by `printf` for `%o` format.

**Features:**
- Treats number as unsigned
- Recursive base-8 digit extraction
- No leading '0' prefix

---

### String/Character Functions

#### `char(string, i)` - Get Character from String
```c
B_TYPE char(B_TYPE string, B_TYPE i)
```
Returns the i-th character of a string (zero-indexed).

**Example:**
```b
s = "Hello";
c = char(s, 0);    /* c = 'H' (72) */
c = char(s, 4);    /* c = 'o' (111) */
```

**Note:**
- Function is named `_char` in C (aliased to `char` in assembly)
- Avoids conflict with C's `char` keyword

**Implementation:** Casts B word pointer to `char*` and indexes.

---

#### `lchar(string, i, chr)` - Set Character in String
```c
void lchar(B_TYPE string, B_TYPE i, B_TYPE chr)
```
Stores a character in the i-th position of a string.

**Example:**
```b
s = "Hello";
lchar(s, 0, 'h');  /* s becomes "hello" */
```

**Implementation:** Casts B word pointer to `char*` and assigns.

---

### File I/O Functions

#### `read()` - Read Character from stdin
```c
B_TYPE read(void)
```
Reads one character from standard input.

**Return Values:**
- `0-127`: ASCII character read successfully
- `4` (ETX): End of file or I/O error
- `0`: Non-ASCII character (filtered out)

**System Call:** `SYS_read` with file descriptor 0 (stdin)

---

#### `nread(file, buffer, count)` - Read Bytes
```c
B_TYPE nread(B_TYPE file, B_TYPE buffer, B_TYPE count)
```
Reads `count` bytes into `buffer` from file descriptor `file`.

**Return Value:**
- Positive: Number of bytes read
- Negative: Error occurred

**System Call:** `SYS_read`

---

#### `nwrite(file, buffer, count)` - Write Bytes
```c
B_TYPE nwrite(B_TYPE file, B_TYPE buffer, B_TYPE count)
```
Writes `count` bytes from `buffer` to file descriptor `file`.

**Return Value:**
- Positive: Number of bytes written
- Negative: Error occurred

**System Call:** `SYS_write`

---

### System Functions

#### `exit()` - Terminate Process
```c
void exit(void)
```
Terminates the current process with exit code 0.

**System Call:** `SYS_exit`

**Note:** B programs automatically return from `main()`, but this function provides explicit termination if needed.

---

#### `flush()` - Flush Output Buffer
```c
void flush(void)
```
Flush output buffer. Currently a no-op (all output is unbuffered).

**Implementation:** Empty function for compatibility with B specifications.

---

## Global Variables

### `fout`
```c
B_TYPE fout = 0;
```
File output descriptor offset. Used by `writeb()` to determine output file:
- `fout = 0` → write to stdout (fd 1)
- `fout = n` → write to fd (n+1)

---

## Compilation

### macOS
```bash
cc -c -ffreestanding runtime/libb.c -o libb.o
```

**Required Flag:** `-ffreestanding` tells the compiler not to assume standard library is available.

### Linux
```bash
gcc -c runtime/libb.c -o libb.o
```

**Note:** The `-ffreestanding` flag is required on macOS but may not be needed on Linux.

---

## Linking with B Programs

After compiling a B program to LLVM IR, link it with `libb.o`:

```bash
# Compile B program
./blang -o program.ll program.b

# Link with runtime library
clang program.ll libb.o -o program

# Run
./program
```

**Linking Options:**
- Standard linking (uses system's C runtime for process startup)
- Can use `-nostdlib` for fully freestanding binaries (requires custom `_start`)

---

## Platform-Specific Details

### macOS
- System call numbers need `0x2000000` offset
- Uses `_char` symbol name with assembly alias to `char`
- Requires `-ffreestanding` compilation flag

### Linux
- Direct system call numbers
- Includes custom `_start` entry point (when compiled for Linux)
- `main()` is called from `_start`, which then calls `SYS_exit`

---

## B Language Integration

### Calling Conventions

B functions receive and return `B_TYPE` (64-bit integers):
```b
/* B code */
x = printf("Value: %d*n", 42);  /* printf returns B_TYPE */
```

```llvm
; LLVM IR
%1 = call i64 (...) @printf(i64 %fmt, i64 42)
```

### Variadic Functions

`printf` is variadic and works with B's calling convention:
```c
void printf(B_TYPE fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    // ... process arguments
    va_end(ap);
}
```

---

## Memory Model

### String Layout

Strings in B are word-aligned character arrays:
```b
s = "Hello";
```

**Memory Layout:**
```
[H][e][l][l][o][\0][padding...]
```

- Null-terminated for C compatibility
- Accessed via `char(s, i)` or direct pointer dereference
- Stored in global constant section

### Multi-Character Constants

Multi-character constants pack characters into a word (big-endian):
```b
c = 'AB';      /* 0x0000000000004142 */
c = 'ABCD';    /* 0x0000000041424344 */
```

The `write()` function unpacks these correctly.

---

## Performance Characteristics

### No Buffering
All I/O is unbuffered and goes directly to system calls:
- **Advantage:** Simple, predictable behavior
- **Disadvantage:** Can be slow for heavy I/O
- **Use Case:** B programs typically do minimal I/O

### Inline Assembly
The system call wrapper uses inline assembly for maximum efficiency:
- Single instruction overhead
- No function call overhead
- Direct register mapping

---

## Limitations and Constraints

### 1. Freestanding Environment
- No standard C library functions (`malloc`, `strlen`, etc.)
- All functionality must use system calls
- Memory allocation not provided (B programs use static allocation)

### 2. Platform Support
- Currently tested on macOS and Linux x86_64
- Requires modification for other architectures
- System call numbers are platform-specific

### 3. Error Handling
- Minimal error handling (most functions return error codes)
- No errno setting
- I/O errors typically ignored or return sentinel values

### 4. Character Encoding
- ASCII only (characters > 127 filtered out)
- No Unicode support
- No locale support

---

## Examples

### Basic Output
```b
main() {
    write('Hello, W');
    write('orld!*n');
}
```

### Formatted Output
```b
main() {
    auto x, y;
    x = 42;
    y = 100;
    printf("x = %d, y = %d*n", x, y);
    printf("Sum: %d*n", x + y);
}
```

### String Manipulation
```b
main() {
    auto s, i, c;
    s = "Hello";
    i = 0;
    while ((c = char(s, i)) != 0) {
        write(c);
        i = i + 1;
    }
    write('*n');
}
```

### File I/O
```b
main() {
    auto fd, buffer[100], n;

    /* Read from stdin (fd 0) */
    n = nread(0, buffer, 100);

    /* Write to stdout (fd 1) */
    nwrite(1, buffer, n);
}
```

---

## Testing

The runtime library is tested through B language test cases:

**Test File:** `compiler_test.go` → `TestLibbFunctions`

**Test Coverage:**
- ✅ `write()` - Multi-character constant output
- ✅ `printf()` - Basic formatting
- ✅ `printf()` - Format specifiers (%d, %c, %s)
- ✅ `char()` - Character extraction
- ✅ Integration with compiled B programs

**Run Tests:**
```bash
go test -v -run TestLibbFunctions
```

---

## Modification Guide

### Adding New Functions

1. **Define function in libb.c:**
   ```c
   B_TYPE B_FN(myfunction)(B_TYPE arg1, B_TYPE arg2) {
       // Implementation
       return result;
   }
   ```

2. **Recompile library:**
   ```bash
   make libb.o
   ```

3. **Use in B programs:**
   ```b
   main() {
       result = myfunction(10, 20);
   }
   ```

### Adding Format Specifiers to printf

Add a new case in the switch statement (line 178):
```c
case 'x':  /* hexadecimal */
    x = va_arg(ap, B_TYPE);
    printhex(x);
    goto loop;
```

### Platform Porting

For new platforms, update:
1. **System call wrapper** - Adjust assembly for target architecture
2. **System call numbers** - Platform-specific values
3. **Conditional compilation** - Add platform-specific sections

---

## Security Considerations

### No Buffer Overflow Protection
Functions like `printf` don't check buffer sizes. B programs are expected to manage memory correctly.

### Direct System Calls
Bypassing the C standard library means:
- No additional security checks
- Direct kernel interaction
- Smaller attack surface (no libc vulnerabilities)

### Type Safety
All B values are untyped 64-bit words:
- No type checking at runtime
- Pointer/integer confusion possible
- Developer responsibility to use correctly

---

## Performance Notes

### Unbuffered I/O
Every `write()` and `writeb()` call results in a system call:
- **Advantage:** Immediate output, simple implementation
- **Disadvantage:** ~1-2μs overhead per character on modern systems
- **Mitigation:** Use `printf` for formatted output, or `nwrite` for bulk I/O

### Recursive Implementations
`write()`, `printd()`, and `printo()` are recursive:
- **Advantage:** Simple, elegant code
- **Disadvantage:** Stack usage grows with number size
- **Practical Impact:** Negligible for typical B programs

---

## Compatibility

### Original B Specification
This library implements the standard B runtime as specified in:
- Ken Thompson's B Tutorial (1972)
- Bell Labs B Language Reference

### Extensions
Some functions may be extensions beyond the original B specification:
- `nread()` and `nwrite()` for binary I/O
- `lchar()` for string modification
- Platform-specific adaptations

### Differences from Original
- Uses 64-bit words (original B used PDP-11 word size)
- System calls instead of OS/360 or Unix V6 primitives
- macOS support (not in original B)

---

## Troubleshooting

### Common Issues

#### 1. Compilation Error: "standard library not found"
**Solution:** Use `-ffreestanding` flag (macOS)
```bash
cc -c -ffreestanding runtime/libb.c -o libb.o
```

#### 2. Linking Error: "undefined reference to printf"
**Solution:** Ensure `libb.o` is included in link command
```bash
clang program.ll libb.o -o program  # Correct
clang program.ll -o program         # Wrong - missing libb.o
```

#### 3. Segmentation Fault in printf
**Possible Causes:**
- String not null-terminated
- Invalid pointer passed as `%s` argument
- Stack corruption from deep recursion

**Debug:** Check that B strings are properly terminated with `\0`

#### 4. Output Appears Garbled
**Causes:**
- Multi-character constants have wrong endianness
- Non-ASCII characters in output
- String encoding issues

**Solution:** Verify character encoding and use ASCII only

---

## Build Integration

### Makefile Target

```makefile
libb.o: runtime/libb.c
	cc -c -ffreestanding runtime/libb.c -o libb.o
```

### Clean Target

```makefile
clean:
	rm -f blang *.o *.ll
```

---

## Future Enhancements

### Potential Additions
- [ ] Memory allocation functions (`alloc`, `free`)
- [ ] More string functions (`strcpy`, `strcmp`)
- [ ] Mathematical functions (`sqrt`, `sin`, `cos`)
- [ ] File operations (`open`, `close`, `seek`)
- [ ] Error reporting utilities
- [ ] Buffered I/O option

### Optimization Opportunities
- [ ] Buffer small writes to reduce syscall overhead
- [ ] Optimize `printd` with iterative algorithm
- [ ] Add SIMD string operations

### Platform Support
- [ ] Windows support (different syscall mechanism)
- [ ] ARM64 support (different assembly)
- [ ] RISC-V support

---

## Technical Specifications

### Function Summary Table

| Function | Purpose | Arguments | Returns | System Call |
|----------|---------|-----------|---------|-------------|
| `write(c)` | Write multi-char | c: packed chars | void | SYS_write |
| `writeb(c)` | Write single byte | c: byte | void | SYS_write |
| `printf(fmt, ...)` | Formatted output | fmt, args | void | via write |
| `printd(n)` | Print decimal | n: number | void | via write |
| `printo(n)` | Print octal | n: number | void | via write |
| `read()` | Read character | none | char/EOF | SYS_read |
| `nread(fd, buf, n)` | Read bytes | fd, buffer, count | bytes read | SYS_read |
| `nwrite(fd, buf, n)` | Write bytes | fd, buffer, count | bytes written | SYS_write |
| `char(s, i)` | Get char | string, index | character | none |
| `lchar(s, i, c)` | Set char | string, index, char | void | none |
| `exit()` | Exit process | none | never returns | SYS_exit |
| `flush()` | Flush output | none | void | none (no-op) |

### Size and Dependencies

**Object File Size:** ~2.6 KB (`libb.o`)
**Dependencies:** None (freestanding)
**System Calls Used:** read, write, exit

---

## License

This runtime library is part of the B language compiler project. See the main project LICENSE file for details.

---

## References

- **Original B Documentation:** [Ken Thompson's B Tutorial](https://www.bell-labs.com/usr/dmr/www/btut.html)
- **B Language Reference:** [CSRC B Manual](https://www.bell-labs.com/usr/dmr/www/kbman.html)
- **System Calls:** `man 2 syscall` on Linux/macOS
- **x86_64 ABI:** [System V AMD64 ABI](https://www.uclibc.org/docs/psABI-x86_64.pdf)

---

**Last Updated:** October 12, 2025
**Maintainer:** B Compiler Project
**Status:** Production-ready, actively maintained
