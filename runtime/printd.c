#include "runtime.h"

//
// The following function will print a decimal number, possibly negative.
// This routine uses the fact that in the ANSCII character set,
// the digits O to 9 have sequential code values.
//
void b_printd(word_t n, ...)
{
    word_t a;

    if (n < 0) {
        b_write('-');
        n = -n;
    }

    if ((a = n / 10))
        b_printd(a);
    b_write(n % 10 + '0');
}
