/* Comprehensive pointer and array test */

global_arr[5] 100, 200, 300, 400, 500;

test_local_arrays() {
    auto arr[3], i, sum;
    arr[0] = 1;
    arr[1] = 2;
    arr[2] = 3;

    sum = 0;
    i = 0;
    while (i < 3) {
        sum = sum + arr[i];
        i++;
    }
    return(sum);  /* 1+2+3 = 6 */
}

test_pointers() {
    auto x, y, ptr;
    x = 42;
    ptr = &x;
    y = *ptr;      /* y = 42 */
    *ptr = 99;     /* x = 99 */
    return(*ptr);  /* returns 99 */
}

test_pointer_arithmetic() {
    auto arr[4], ptr;
    arr[0] = 10;
    arr[1] = 20;
    arr[2] = 30;
    arr[3] = 40;

    ptr = &arr[0];
    return(*(ptr + 2));  /* Should be arr[2] = 30 */
}

test_global_arrays() {
    extrn global_arr;
    return(global_arr[2]);  /* Should be 300 */
}

main() {
    auto result;

    result = test_local_arrays();      /* 6 */
    result = test_pointers();          /* 99 */
    result = test_pointer_arithmetic(); /* 30 */
    result = test_global_arrays();     /* 300 */

    return(result);
}
