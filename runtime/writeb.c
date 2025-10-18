#include "runtime.h"

//
// Select output stream: 0-stdout, 1-stderr.
//
word_t b_fout = 0;

//
// One byte is written on the standard output file.
//
void b_writeb(word_t c, ...)
{
    syscall(SYS_write, b_fout + 1, (word_t)&c, 1);
}
