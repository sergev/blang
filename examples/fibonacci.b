/* This program prints out the n-th fibonacci number. */

n 10;

main() {
    extrn n;
    printf("%d*n", fib(n));
}

fib(n) {
    auto a, b, c, i;
    b = 1;
    i = 0;

    while (i < n) {
        c = a + b;
        a = b;
        b = c;
        i = i + 1;
    }

    return(a);
}
