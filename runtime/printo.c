#include "runtime.h"

//
// The following function will print an unsigned number, n,
// to the base 8.
//
void printo(word_t n)
{
    word_t a;

    if ((a = (uintptr_t)n >> 3))
        printo(a);
    write((n & 7) + '0');
}
