#include "runtime.h"

//
// Count bytes are read into the vector buffer from the open
// file designated by file. The actual number of bytes read
// are returned. A negative number returned indicates an error.
//
word_t nread(word_t file, word_t buffer, word_t count)
{
    return (word_t)syscall(SYS_read, file, buffer, count);
}
