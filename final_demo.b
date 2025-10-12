/* Demonstration of B Compiler with LLVM Backend */

/* Recursive Fibonacci */
fib(n) {
    if (n <= 1)
        return(n);
    return(fib(n - 1) + fib(n - 2));
}

/* Test all comparison operators */
compare(a, b) {
    auto result;
    result = 0;

    if (a < b) result = result + 1;
    if (a <= b) result = result + 2;
    if (a > b) result = result + 4;
    if (a >= b) result = result + 8;
    if (a == b) result = result + 16;
    if (a != b) result = result + 32;

    return(result);
}

/* Test operator precedence */
precedence_test() {
    auto x;

    x = 2 + 3 * 4;           /* = 14 (mult first) */
    x = (2 + 3) * 4;         /* = 20 (parens first) */
    x = 10 & 7 | 2;          /* = (10 & 7) | 2 = 2 | 2 = 2 */
    x = 5 << 2 >> 1;         /* = (5 << 2) >> 1 = 20 >> 1 = 10 */

    return(x);
}

/* Main entry point */
main() {
    auto result;

    result = fib(10);              /* Fibonacci(10) = 55 */
    result = compare(10, 20);      /* Should be 35 (1+2+32) */
    result = precedence_test();    /* Should be 10 */

    return(result);
}
