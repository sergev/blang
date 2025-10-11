/* Test various B language features */

/* Global variables */
global_var 42;
global_array[5] 1, 2, 3, 4, 5;

/* Function with arguments and local variables */
add(x, y)
{
    auto result;
    result = x + y;
    return(result);
}

/* Function with control flow */
factorial(n)
{
    if (n <= 1)
        return(1);
    return(n * factorial(n - 1));
}

/* Main function */
main()
{
    extrn global_array;
    auto i, sum;

    /* Test arithmetic */
    sum = add(10, 20);

    /* Test loops */
    i = 0;
    while (i < 5) {
        sum = sum + global_array[i];
        i++;
    }

    /* Test conditionals */
    if (sum > 100) {
        sum = 100;
    } else {
        sum = 0;
    }

    return(sum);
}
