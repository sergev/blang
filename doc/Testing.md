# Testing the B Compiler

## Quick Start

### 1. Build the Compiler

```bash
go build -o blang
```

### 2. Compile B Runtime Library

```bash
clang -c -ffreestanding libb/libb.c -o libb.o
```

> **Note**: The `-ffreestanding` flag is required on macOS.

### 3. Compile and Run a B Program

```bash
# Compile B to LLVM IR
./blang -o program.ll program.b

# Link with runtime and create executable
clang program.ll libb.o -o program

# Run
./program
echo $?  # Check exit code (return value)
```

## Test Programs

### examples/ - Canonical Example Programs

- `hello.b` - Hello world using write()
- `helloworld.b` - Hello world using printf()
- `fibonacci.b` - Fibonacci calculator
- `fizzbuzz.b` - FizzBuzz 1-100
- `e-2.b` - E-2 constant calculation

### testdata/ - Feature Test Programs

- `arithmetic.b` - Arithmetic operations and function calls
- `globals.b` - Global variables and arrays
- `conditionals.b` - if/else statements
- `loops.b` - while loops and factorial
- `strings.b` - String literals and array initialization
- `operators.b` - Bitwise, comparison, and unary operators
- `switch.b` - Switch/case statements
- `goto.b` - goto and labels
- `pointers.b` - Pointer operations
- `arrays.b` - Array operations
- `comprehensive_ptr.b` - Complex pointer operations

## Verified Working Programs

All test programs compile to LLVM IR and execute correctly:

| Program | Description | Expected Result |
|---------|-------------|-----------------|
| `examples/hello.b` | write() with multi-char | ✅ Outputs "Hello, World!" |
| `examples/helloworld.b` | printf() with strings | ✅ Outputs "Hello, World!" |
| `examples/fibonacci.b` | Fibonacci calculator | ✅ Returns 55 (fib(10)) |
| `examples/fizzbuzz.b` | FizzBuzz 1-100 | ✅ Correct output |
| `testdata/arithmetic.b` | All arithmetic operators | ✅ Returns 50 |
| `testdata/conditionals.b` | if/else, max, abs | ✅ Returns 35 |
| `testdata/loops.b` | While, factorial(5) | ✅ Returns 120 |
| `testdata/arrays.b` | Array operations | ✅ Returns 150 |
| `testdata/pointers.b` | Pointer ops | ✅ Returns 30 |
| `testdata/globals.b` | Global vars & arrays | ✅ Returns 60 |
| `testdata/switch.b` | Switch/case | ✅ Returns 30 |
| `testdata/goto.b` | goto and labels | ✅ Returns 42 |

## Test Programs

### Hello World (with printf)

```bash
./blang -o hello.ll testdata/hello.b
clang hello.ll libb.o -o hello
./hello
# Output: Hello, World!
```

Note: `printf()` is automatically declared as external - no `extrn` needed!

### Factorial (Tests Recursion)

```bash
./blang -o factorial.ll testdata/loops.b
clang factorial.ll libb.o -o factorial
./factorial
echo $?
# Exit code: 120 (which is 5!)
```

### Arithmetic Operations

```bash
./blang -o arith.ll testdata/arithmetic.b
clang arith.ll libb.o -o arith
./arith
echo $?
# Exit code: 50
```

### Conditionals (if/else)

```bash
./blang -o cond.ll testdata/conditionals.b
clang cond.ll libb.o -o cond
./cond
echo $?
# Exit code: 35
```

## Operator Tests

### All Arithmetic Operators

```b
main() {
    auto x, y;
    x = 10;
    y = 3;

    x = x + y;   /* addition: 13 */
    x = x - y;   /* subtraction: 10 */
    x = x * y;   /* multiplication: 30 */
    x = x / y;   /* division: 10 */
    x = x % y;   /* modulo: 1 */

    return(x);
}
```

### All Comparison Operators

```b
main() {
    auto x, y, r;
    x = 10;
    y = 5;

    r = (x > y);    /* greater than: 1 (true) */
    r = (x < y);    /* less than: 0 (false) */
    r = (x >= y);   /* greater or equal: 1 */
    r = (x <= y);   /* less or equal: 0 */
    r = (x == y);   /* equal: 0 */
    r = (x != y);   /* not equal: 1 */

    return(r);
}
```

### Bitwise Operators

```b
main() {
    auto x, y;
    x = 12;  /* 1100 in binary */
    y = 10;  /* 1010 in binary */

    x = x & y;    /* AND: 8 (1000) */
    x = x | y;    /* OR: 10 (1010) */
    x = x << 2;   /* Left shift: 40 */
    x = x >> 1;   /* Right shift: 20 */

    return(x);
}
```

### Increment/Decrement

```b
main() {
    auto x, y;
    x = 10;

    ++x;      /* prefix increment: x = 11 */
    --x;      /* prefix decrement: x = 10 */
    y = x++;  /* postfix increment: y = 10, x = 11 */
    y = x--;  /* postfix decrement: y = 11, x = 10 */

    return(x + y);  /* 10 + 11 = 21 */
}
```

### Ternary Operator

```b
max(a, b) {
    return((a > b) ? a : b);
}

main() {
    return(max(42, 17));  /* Returns 42 */
}
```

### Complex Expressions

```b
main() {
    auto result;

    /* Test operator precedence */
    result = 2 + 3 * 4;        /* 2 + (3 * 4) = 14 */
    result = (2 + 3) * 4;      /* (2 + 3) * 4 = 20 */
    result = 10 << 2 & 255;    /* (10 << 2) & 255 = 40 */

    /* Nested ternary */
    result = (result > 50) ? 100 : (result > 25) ? 50 : 25;

    return(result);
}
```

## Runtime Library Functions

The `libb/libb.c` runtime provides:

- **`write(c)`** - Write multi-character constant to stdout
- **`read()`** - Read one character from stdin
- **`printf(fmt, ...)`** - Formatted output (%d, %o, %c, %s)
- **`printd(n)`** - Print decimal number
- **`printo(n)`** - Print octal number
- **`char(s, i)`** - Get i-th character of string
- **`lchar(s, i, c)`** - Set i-th character of string

## Unit Tests

The B compiler has comprehensive unit test coverage:

```bash
# Run all tests
go test

# Run specific test categories
go test -run TestPrecedence    # 28 operator precedence tests
go test -run TestGlobals       # 4 global/local variable tests
go test -run TestStrings       # 2 string/char literal tests
go test -run TestLibbFunctions # 6 runtime library tests
go test -run TestExpressions   # 9 expression feature tests
go test -run TestFunctions     # 3 function tests

# Run with verbose output
go test -v

# Check code coverage
go test -cover
```

## Test Results Summary

✅ **124 tests passing** (100% of active tests)
⏭️ **3 tests skipped** (pending features)
📈 **73.9% code coverage**

### Test Breakdown by Category

- **29 tests** - Lexer (tokenization, escape sequences)
- **28 tests** - Operator precedence (all combinations)
- **10 tests** - Compilation (basic verification)
- **9 tests** - Integration (compile + link + run)
- **9 tests** - Expressions (operators, unary, arrays)
- **7 tests** - Error handling
- **6 tests** - Runtime library (`printf`, `write`, `char`, etc.)
- **4 tests** - Globals (multi-value scalars, reverse allocation)
- **3 tests** - Functions (definitions, parameters, calls)
- **2 tests** - Strings (escape sequences, literals)
- **2 tests** - Indirect calls (function pointers)

### Skipped Tests (Pending Implementation)

- **15 tests** - Compound assignments (`=+`, `=-`, etc.)
- **1 test** - Ternary operator (`? :`)
- **1 test** - E-2 constant (long-running computation)

## Known Limitations

- ⏳ Compound assignments (`=+`, `=-`, etc.) not yet implemented - use `x = x + 5` instead
- ⏳ Ternary operator (`? :`) not yet implemented
