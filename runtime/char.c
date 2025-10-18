#include "runtime.h"

//
// The i-th character of the string is returned.
//
word_t b_char(word_t string, /*word_t i,*/ ...)
{
    va_list ap;
    va_start(ap, string);
    word_t i = va_arg(ap, word_t);
    va_end(ap);

    return ((char *)string)[i];
}
