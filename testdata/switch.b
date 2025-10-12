/* Test switch/case statements */

classify(n) {
    auto result;
    result = -1;

    switch (n) {
        case 0:
            result = 0;
        case 1:
            result = 10;
        case 2:
            result = 20;
        case 3:
            result = 30;
    }

    return(result);
}

main() {
    auto r;
    r = classify(0);  /* Should be 0 */
    r = classify(2);  /* Should be 20 */
    r = classify(3);  /* Should be 30 */
    return(r);
}
