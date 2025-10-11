/* Test global variables */

counter 0;
values[3] 10, 20, 30;

increment() {
    extrn counter;
    counter = counter + 1;
}

sum_values() {
    extrn values;
    auto i, total;
    total = 0;
    i = 0;
    while (i < 3) {
        total = total + values[i];
        i++;
    }
    return(total);
}

main() {
    increment();
    increment();
    return(sum_values());
}

