#include "runtime.h"

//
// The character char is stored in the i-th character of the string.
//
void b_lchar(word_t string, word_t i, word_t chr, ...)
{
    ((char *)string)[i] = chr;
}
