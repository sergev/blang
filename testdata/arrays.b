/* Test array operations */

sum_array(arr, n) {
    auto i, sum;
    sum = 0;
    i = 0;
    while (i < n) {
        sum = sum + arr[i];
        i++;
    }
    return(sum);
}

main() {
    auto numbers[5];
    auto i, total;

    /* Initialize array */
    i = 0;
    while (i < 5) {
        numbers[i] = (i + 1) * 10;
        i++;
    }

    /* Sum using function */
    total = sum_array(numbers, 5);

    return(total);  /* Should be 10+20+30+40+50 = 150 */
}
