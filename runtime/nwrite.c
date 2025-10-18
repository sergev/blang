#include "runtime.h"

//
// Count bytes are written out of the vector buffer on the
// open file designated by file. The actual number of bytes
// written are returned. A negative number returned indicates an error.
//
word_t b_nwrite(word_t file, /*word_t buffer, word_t count,*/ ...)
{
    va_list ap;
    va_start(ap, file);
    word_t buffer = va_arg(ap, word_t);
    word_t count  = va_arg(ap, word_t);
    va_end(ap);

    return (word_t)syscall(SYS_write, file, buffer, count);
}
