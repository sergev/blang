#include "runtime.h"

//
// Select output stream: 0-stdout, 1-stderr.
//
word_t fout = 0;

//
// One byte is written on the standard output file.
//
void writeb(word_t c)
{
    syscall(SYS_write, fout + 1, (word_t)&c, 1);
}
