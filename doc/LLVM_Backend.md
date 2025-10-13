# LLVM Backend Implementation

## Overview

The B compiler generates LLVM IR instead of x86_64 assembly, providing portability across all LLVM-supported architectures.

**Status:** Complete implementation with full B language support

## Implementation Details

### Files

- **`codegen.go`** - LLVM IR generation infrastructure
- **`parser.go`** - B language parser for LLVM backend
- **`expr.go`** - Comprehensive expression parser with full operator precedence
- **`control.go`** - Control flow statement handling (switch/case, goto)

**Total:** Complete B language implementation with LLVM backend

### Expression Parser

The expression parser implements all 15 precedence levels from the B language:

| Level | Operators | Associativity | Example |
|-------|-----------|---------------|---------|
| 15 | Primary, Postfix `()` `[]` `++` `--` | Left | `a[i]`, `f()`, `x++` |
| 14 | Assignment `=` `=+` `=-` etc | Right | `x = y`, `x =+ 5` |
| 13 | Ternary `?:` | Right | `x ? y : z` |
| 10 | Bitwise OR `\|` | Left | `a \| b` |
| 8 | Bitwise AND `&` | Left | `a & b` |
| 7 | Equality `==` `!=` | Left | `a == b` |
| 6 | Relational `<` `<=` `>` `>=` | Left | `a < b` |
| 5 | Shift `<<` `>>` | Left | `a << 2` |
| 4 | Additive `+` `-` | Left | `a + b` |
| 3 | Multiplicative `*` `/` `%` | Left | `a * b` |
| 2 | Unary `!` `-` `++` `--` `*` `&` | Right | `!x`, `-y`, `&z` |
| 1 | Primary | - | Literals, identifiers, `()` |

### Automatic External Function Declaration

Following B language semantics, undefined identifiers used in function call position are automatically declared as external functions:

```b
main() {
    write('Hello');   /* write() automatically declared as external */
    printf("Hi %d", 42);  /* printf() automatically declared as external */
}
```

This allows B programs to call C library functions **without explicit `extrn` declarations**. No need to write `extrn write, printf;` - the compiler handles it automatically!

### Function Call Semantics

**Direct Function Calls (Recommended):**
```b
printf("hello");  /* Undefined function, auto-declared as external ✅ */
```

**Indirect Function Calls (via Variables):**
```b
add(a, b) { return(a + b); }
func_ptr;               /* Global variable */

main() {
    extrn func_ptr;     /* Declares func_ptr as external variable */
    func_ptr = add;     /* Function name is its address */
    printf("%d", func_ptr(3, 5));  /* Indirect call through pointer ✅ */
}
```

**Note:** Function names are automatically their addresses in B (no `&` operator needed).

## Compiling and Linking

### Compile B to LLVM IR

```bash
./blang -o program.ll program.b
```

### Compile and Link

```bash
# First time: compile the B runtime library (do this once)
clang -c -ffreestanding libb/libb.c -o libb.o

# Compile B program to LLVM IR
./blang -o program.ll program.b

# Link with runtime and create executable
clang program.ll libb.o -o program

# Run
./program
```

### Complete Example

```bash
# Compile factorial program
./blang -o factorial.ll testdata/loops.b

# Link with runtime
clang factorial.ll libb.o -o factorial

# Run (calculates 5! = 120)
./factorial
echo $?  # Prints 120
```

## Runtime Library (libb.c)

The B runtime library provides:

- **`write(c)`** - Write multi-character constant to stdout (not string pointers!)
- **`printf(fmt, ...)`** - Formatted output for strings (%d, %o, %c, %s, %%)
- **`printd(n)`** - Print decimal number
- **`printo(n)`** - Print octal number
- **`read()`** - Read character from stdin
- **`char(s, i)`** - Get i-th character of string
- **`lchar(s, i, c)`** - Set i-th character of string
- **`exit()`** - Terminate program
- **`flush()`** - Flush output (no-op)

### Important: write() vs printf()

- **`write('Hello')`** - For multi-character constants (packed into i64)
- **`printf("Hello")`** - For string pointers (null-terminated char arrays)

Example:
```b
main() {
    extrn write, printf;
    write('Hi!*n');           /* Multi-char constant */
    printf("Hello %s*n", "World");  /* String pointer */
}
```

## Advanced Features

### Scalar with Multiple Initialization Values

B allows allocating multiple consecutive words for a scalar variable:

```b
c -345, 'foo', "bar";  /* Allocates 3 words */

main() {
    auto ptr;

    ptr = &c;
    printf("%d*n", c);      /* Prints -345 (first value) */
    printf("%c*n", ptr[1]); /* Prints 'foo' (second value) */
    printf("%s*n", ptr[2]); /* Prints "bar" (third value) */
}
```

**Memory Layout:**
- `&c + 0`: -345
- `&c + 8`: 'foo' (0x666F6F)
- `&c + 16`: pointer to "bar"

### Auto Arrays with Expression Sizes

Array sizes can be numeric or character constants:

```b
main() {
    auto buffer['x'];     /* Size = 120 (ASCII 'x') */
    auto data[100];       /* Size = 100 */
    auto temp[];          /* Size = 1 (default) */
}
```

### Reverse Allocation Order

Auto statements are allocated in reverse order on the stack:

```b
main() {
    auto a, b;     /* Statement 1 */
    auto c[10];    /* Statement 2 */
    auto d, e;     /* Statement 3 */

    /* Stack allocation: d, e → c → a, b */
}
```

Variables within one statement are allocated in forward order.

### Indirect Function Calls

Function pointers can be stored in variables and called indirectly:

```b
add(a, b) { return(a + b); }
sub(a, b) { return(a - b); }

operation;  /* Global variable to hold function pointer */

main() {
    extrn operation;

    operation = add;
    printf("add: %d*n", operation(10, 5));  /* Prints: add: 15 */

    operation = sub;
    printf("sub: %d*n", operation(10, 5));  /* Prints: sub: 5 */
}
```

**Key Points:**
- Function names are automatically their addresses (no `&` needed)
- Variables hold function pointers as `i64` values
- Calls through variables use LLVM indirect call mechanism
- Supports variadic function pointers

## Testing

All test programs compile successfully to LLVM IR:

```bash
# Test arithmetic operations
./blang -o arithmetic.ll testdata/arithmetic.b

# Test conditionals (if/else)
./blang -o conditionals.ll testdata/conditionals.b

# Test loops (while, factorial)
./blang -o loops.ll testdata/loops.b

# Test all operators
./blang -o operators.ll testdata/operators.b

# Run unit tests (120 tests)
go test
```

## Verified Features

✅ **Operator Precedence** - All 15 levels implemented correctly (28 tests)
✅ **Type Conversions** - Boolean (i1) → Integer (i64) automatic conversion
✅ **Lvalue/Rvalue** - Proper handling of addressable vs value expressions
✅ **Function Calls** - Including automatic external function declaration
✅ **Control Flow** - if/else, while loops, switch/case, goto, labels
✅ **Recursion** - Tested with factorial and fibonacci
✅ **Complex Expressions** - Nested operators, chained comparisons, parentheses
✅ **Scalar with Multiple Values** - `c -345, 'foo', "bar";` allocates consecutive words
✅ **Auto Arrays** - Character constant sizes supported: `auto buf['x'];`
✅ **Reverse Allocation** - Auto statements allocated in reverse order
✅ **Escape Sequences** - All B escape sequences verified (10 sequences)
✅ **Indirect Function Calls** - Function pointers stored in variables and called through them

## Example Generated IR

**Input (factorial.b):**
```b
factorial(n) {
    if (n <= 1)
        return(1);
    return(n * factorial(n - 1));
}
```

**Output (LLVM IR):**
```llvm
define i64 @factorial(i64 %n) {
entry:
  %0 = alloca i64
  store i64 %n, i64* %0
  %1 = load i64, i64* %0
  %2 = icmp sle i64 %1, 1
  %3 = zext i1 %2 to i64
  %4 = icmp ne i64 %3, 0
  br i1 %4, label %if.then, label %if.else

if.then:
  ret i64 1

if.else:
  br label %if.end

if.end:
  %5 = load i64, i64* %0
  %6 = load i64, i64* %0
  %7 = sub i64 %6, 1
  %8 = call i64 @factorial(i64 %7)
  %9 = mul i64 %5, %8
  ret i64 %9
}
```

Clean, optimizable LLVM IR that can be compiled to any LLVM-supported architecture!
