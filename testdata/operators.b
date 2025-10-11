/* Test various operators */

test_bitwise(a, b) {
    auto x, y, z;
    x = a & b;
    y = a | b;
    z = a << 2;
    return(x + y + z);
}

test_comparison(a, b) {
    if (a == b)
        return(0);
    if (a != b)
        return(1);
    return(2);
}

test_unary() {
    auto x;
    x = 10;
    x++;
    ++x;
    x--;
    --x;
    return(x);
}

main() {
    auto x;
    x = test_bitwise(12, 5);
    x = x + test_comparison(10, 20);
    x = x + test_unary();
    return(x);
}

