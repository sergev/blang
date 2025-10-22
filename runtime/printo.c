#include "runtime.h"

//
// The following function will print an unsigned number, n,
// to the base 8.
//
void b_printo(word_t n, ...)
{
    uintptr_t value = (uintptr_t)n;
    char buf[(sizeof(uintptr_t) * 8 + 2) / 3];
    char *end = buf + sizeof(buf);
    char *p = end;

    do {
        *--p = '0' + (value & 7);
        value >>= 3;
    } while (value != 0);

    b_nwrite(b_fout + 1, (word_t)p, (word_t)(end - p));
}
