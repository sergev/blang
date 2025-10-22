#include "runtime.h"

//
// The following function will print a decimal number, possibly negative.
//
void b_printd(word_t n, ...)
{
    uintptr_t value = (uintptr_t)n;
    int negative = n < 0;
    char buf[2 + sizeof(word_t) * 3];
    char *end = buf + sizeof(buf);
    char *p = end;

    if (negative) {
        value = 1 + ~value; // avoid overflow on MIN
    }

    do {
        *--p = '0' + (char)(value % 10);
        value /= 10;
    } while (value != 0);

    if (negative) {
        *--p = '-';
    }

    b_nwrite(b_fout + 1, (word_t)p, (word_t)(end - p));
}
