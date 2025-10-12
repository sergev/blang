/* Test goto and labels */

main() {
    auto x, y;
    x = 0;
    y = 0;

    /* Jump over some code */
    goto skip;
    x = 100;  /* This should be skipped */

skip:
    y = 42;

    /* Conditional goto */
    if (y > 40)
        goto done;

    x = 200;  /* This should be skipped */

done:
    return(y);  /* Should return 42 */
}
