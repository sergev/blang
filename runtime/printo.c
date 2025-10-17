#include "runtime.h"

//
// The following function will print an unsigned number, n,
// to the base 8.
//
void b_printo(word_t n, ...)
{
    word_t a;

    if ((a = (uintptr_t)n >> 3))
        b_printo(a);
    b_write((n & 7) + '0');
}
