/* B Language Showcase - All Features */

/* Global array */
fibonacci_cache[20] 0, 1;

/* Initialize fibonacci cache */
init_cache() {
    extrn fibonacci_cache;
    auto i;
    i = 2;
    while (i < 20) {
        fibonacci_cache[i] = fibonacci_cache[i-1] + fibonacci_cache[i-2];
        i++;
    }
}

/* Get nth fibonacci number from cache */
fib(n) {
    extrn fibonacci_cache;
    if (n >= 20)
        return(-1);  /* Error */
    return(fibonacci_cache[n]);
}

/* Test all operators */
test_operators(a, b) {
    auto result;

    /* Arithmetic */
    result = a + b - 5;
    result = a * b / 10;
    result = a % b;

    /* Bitwise */
    result = a & b;
    result = a | b;
    result = a << 2;
    result = a >> 1;

    /* Comparison */
    result = (a > b) ? a : b;
    result = (a == b);
    result = (a != b);

    /* Unary */
    result = -a;
    result = !result;

    return(result);
}

/* Test pointers and arrays */
test_arrays() {
    auto arr[5], i, sum, ptr;

    /* Initialize */
    i = 0;
    while (i < 5) {
        arr[i] = (i + 1) * 10;
        ++i;
    }

    /* Sum using indexing */
    sum = 0;
    i = 0;
    while (i < 5) {
        sum = sum + arr[i];
        i++;
    }

    /* Access via pointer */
    ptr = arr;
    sum = sum + ptr[2];  /* Add arr[2] again */

    return(sum);  /* 10+20+30+40+50+30 = 180 */
}

/* Main entry point */
main() {
    auto result;

    /* Initialize fibonacci cache */
    init_cache();

    /* Print some fibonacci numbers */
    printf("Fibonacci(5) = %d*n", fib(5));
    printf("Fibonacci(10) = %d*n", fib(10));

    /* Test operators */
    result = test_operators(15, 7);
    printf("Operator test result: %d*n", result);

    /* Test arrays */
    result = test_arrays();
    printf("Array test result: %d*n", result);

    return(0);
}
