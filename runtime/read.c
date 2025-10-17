#include "runtime.h"

//
// The next character form the standard input file is returned.
// The character ‘*e’ is returned for an end-of-file.
//
word_t read()
{
    char c = 0;

    if (syscall(SYS_read, 0, (word_t)&c, 1) == 1) {
        if (c > 0 && c <= 127) {
            return c;
        } else {
            // Non-ascii character.
            return 0;
        }
    } else {
        // End of file or i/o error.
        return 4; // ETX
    }
}
