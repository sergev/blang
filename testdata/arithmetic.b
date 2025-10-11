/* Test arithmetic operations */

add(a, b) {
    return(a + b);
}

sub(a, b) {
    return(a - b);
}

mul(a, b) {
    return(a * b);
}

main() {
    auto x, y, z;
    x = 10;
    y = 20;
    z = add(x, y);
    z = sub(z, 5);
    z = mul(z, 2);
    return(z);
}

