#include "runtime.h"

//
// The i-th character of the string is returned.
//
word_t _char(word_t string, word_t i)
{
    return ((char *)string)[i];
}
