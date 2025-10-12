/* Test pointer operations */

main() {
    auto x, y, ptr;

    /* Test address-of and dereference */
    x = 42;
    ptr = &x;        /* Get address of x */
    y = *ptr;        /* Dereference ptr, should be 42 */
    *ptr = 100;      /* Store through pointer */

    /* Test pointer arithmetic */
    auto array[5];
    auto i;

    i = 0;
    while (i < 5) {
        array[i] = i * 10;
        i++;
    }

    /* Access through indexing */
    y = array[2];    /* Should be 20 */

    /* Access through pointer using [] operator */
    ptr = array;     /* In B, array name gives pointer value */
    y = ptr[3];      /* Should be 30 */

    return(y);
}
