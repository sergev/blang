/* Test if/else statements */

max(a, b) {
    if (a > b)
        return(a);
    else
        return(b);
}

abs(n) {
    if (n < 0)
        return(-n);
    return(n);
}

main() {
    auto x, y;
    x = max(10, 20);
    y = abs(-15);
    return(x + y);
}

