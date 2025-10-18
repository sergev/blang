#include "runtime.h"

//
// The character char is stored in the i-th character of the string.
//
void b_lchar(word_t string, /*word_t i, word_t chr,*/ ...)
{
    va_list ap;
    va_start(ap, string);
    word_t i   = va_arg(ap, word_t);
    word_t chr = va_arg(ap, word_t);
    va_end(ap);

    ((char *)string)[i] = chr;
}
