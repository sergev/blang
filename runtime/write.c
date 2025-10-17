#include "runtime.h"

//
// One or more characters are written on the standard output file.
//
void b_write(word_t ch)
{
    char buf[sizeof(word_t)];
    char *p = buf;
    uintptr_t input = ch;
    unsigned len;

    for (len = 0; len < sizeof(word_t); len++, input <<= 8) {
        uint8_t byte = input >> ((sizeof(word_t) - 1) * 8);

        if (byte != 0 || p != buf || len == sizeof(word_t) - 1) {
            *p++ = byte;
        }
    }
    syscall(SYS_write, b_fout + 1, (word_t)buf, p - buf);
}
