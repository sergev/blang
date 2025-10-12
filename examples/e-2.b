/* The following program will calculate the constant e-2 to about
   4000 decimal digits, and print it 50 characters to the line in
   groups of 5 characters.  The method is simple output conversion
   of the expansion
     1/2! + 1/3! + ... = .111....
   where the bases of the digits are 2, 3, 4, . . . */

main() {
    extrn n, v;
    auto i, c, col, a;

    i = col = 0;
    while (i < n) {
        v[i] = 1;
        i = i + 1;
    }
    while (col < 2 * n) {
        a = n + 1;
        c = i = 0;
        while (i < n) {
            c = c + v[i] * 10;
            v[i] = c % a;
            c = c / a;
            a = a - 1;
            i = i + 1;
        }

        write(c + '0');
        col = col + 1;
        if (!(col % 5)) {
            if (col % 50)
                write(' ');
            else
                write('*n');
        }
    }
    write('*n*n');
}
v[2000];
n 2000;
