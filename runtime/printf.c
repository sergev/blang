#include "runtime.h"

//
// The following function is a general formatting, printing, and
// conversion subroutine. The first argument is a format string.
// Character sequences,of the form ‘%x’ are interpreted and cause
// conversion of type x’ of the next argument, other character
// sequences are printed verbatim.
//
void b_printf(word_t fmt, ...)
{
    word_t x, c, i = 0, j;

    va_list ap;
    va_start(ap, fmt);
loop:
    while ((c = b_char(fmt, i++)) != '%') {
        if (c == '\0')
            goto end;
        b_write(c);
    }
    switch (c = b_char(fmt, i++)) {
    case 'd': // decimal
        x = va_arg(ap, word_t);
        if (x < 0) {
            x = -x;
            b_write('-');
        }
        b_printd(x);
        goto loop;

    case 'o': // octal
        x = va_arg(ap, word_t);
        if (x < 0) {
            x = -x;
            b_write('-');
        }
        b_printo(x);
        goto loop;

    case 'c':
        x = va_arg(ap, word_t);
        b_write(x);
        goto loop;

    case 's':
        x = va_arg(ap, word_t);
        j = 0;
        while ((c = b_char(x, j++)) != '\0')
            b_write(c);
        goto loop;
    case '%':
        b_write('%');
        goto loop;
    }
    // Unknown format.
    b_write('%');
    i--;
    goto loop;

end:
    va_end(ap);
}
