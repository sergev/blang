/* Test while loops */

factorial(n) {
    auto result, i;
    result = 1;
    i = 1;
    while (i <= n) {
        result = result * i;
        i++;
    }
    return(result);
}

main() {
    return(factorial(5));
}

