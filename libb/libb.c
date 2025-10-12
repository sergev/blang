//
// Implement read(), write(c), flush().
//
#include <stdarg.h>
#ifndef B_TYPE
    /* type representing B's single data type (64-bit int on x86_64) */
    #include <stdint.h>
    #define B_TYPE intptr_t
#endif
#ifndef B_FN
    /* this macro allows to give each B std function a pre- or postfix
       to avoid issues with common names
       e.g.: #define B_FN(name) __b##name <-- gives each B funcion the prefix "__b" */
    #define B_FN(name) name
#endif

//
// Syscall wrapper implementation
//

/* type used for syscalls */
#define SYSCALL_TYPE long

static inline SYSCALL_TYPE syscall1(SYSCALL_TYPE n, SYSCALL_TYPE a1)
{
	unsigned SYSCALL_TYPE ret;
	__asm__ __volatile__ ("syscall" : "=a"(ret) : "a"(n), "D"(a1) : "rcx", "r11", "memory");
	return ret;
}

static inline SYSCALL_TYPE syscall3(SYSCALL_TYPE n, SYSCALL_TYPE a1, SYSCALL_TYPE a2, SYSCALL_TYPE a3)
{
	unsigned SYSCALL_TYPE ret;
	__asm__ __volatile__ ("syscall" : "=a"(ret) : "a"(n), "D"(a1), "S"(a2),
						  "d"(a3) : "rcx", "r11", "memory");
	return ret;
}

/* syscall ids */
#define SYS_read 0
#define SYS_write 1
#define SYS_open 2
#define SYS_close 3
#define SYS_stat 4
#define SYS_fstat 5
#define SYS_seek 8
#define SYS_fork 57
#define SYS_execve 59
#define SYS_exit 60
#define SYS_wait4 61
#define SYS_chdir 80
#define SYS_mkdir 83
#define SYS_creat 85
#define SYS_link 86
#define SYS_unlink 87
#define SYS_chmod 90
#define SYS_chown 92
#define SYS_gettimeofday 96
#define SYS_getuid 102
#define SYS_setuid 105
#define SYS_time 201

//
// B standard library implementation
//

/* The `main` function must be declared in any B program */
extern B_TYPE B_FN(main)(void);

/* entry point of any B program */
void _start(void) __asm__ ("_start"); /* assure, that _start is really named _start in asm */
void _start(void) {
    B_TYPE code = B_FN(main)();
    syscall1(SYS_exit, code);
}

/* The current process is terminated. */
void B_FN(exit)(void) {
    syscall1(SYS_exit, 0);
}

/* The i-th character of the string is returned */
B_TYPE B_FN(_char)(B_TYPE string, B_TYPE i) __asm__ ("char"); /* alias name */
B_TYPE B_FN(_char)(B_TYPE string, B_TYPE i) {
    return ((char*) string)[i];
}

/* The character char is stored in the i-th character of the string. */
void B_FN(lchar)(B_TYPE string, B_TYPE i, B_TYPE chr) {
    ((char*) string)[i] = chr;
}

//
// The next character form the standard input file is returned.
// The character ‘*e’ is returned for an end-of-file.
//
B_TYPE B_FN(read)(void) {
    B_TYPE c = 0;
    if (syscall3(SYS_read, 0, (B_TYPE)&c, 1) == 1) {
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

/* Count bytes are read into the vector buffer from the open
   file designated by file. The actual number of bytes read
   are returned. A negative number returned indicates an
   error. */
B_TYPE B_FN(nread)(B_TYPE file, B_TYPE buffer, B_TYPE count) {
    return (B_TYPE) syscall3(SYS_read, file, buffer, count);
}

//
// One or more characters are written on the standard output file.
//
B_TYPE fout = 0;

//
// One byte is written on the standard output file.
//
void B_FN(writeb)(B_TYPE c) {
    syscall3(SYS_write, fout + 1, (B_TYPE)&c, 1);
}

//
// One or more characters are written on the standard output file.
//
void B_FN(write)(B_TYPE c) {
    B_TYPE a;

    if ((a = (uintptr_t)c >> 8))
        B_FN(write)(a);

    writeb(c);
}

/* Count bytes are written out of the vector buffer on the
   open file designated by file. The actual number of bytes
   written are returned. A negative number returned indicates
   an error. */
B_TYPE B_FN(nwrite)(B_TYPE file, B_TYPE buffer, B_TYPE count) {
    return (B_TYPE) syscall3(SYS_write, file, buffer, count);
}

/* The following function will print a decimal number, possibly negative.
   This routine uses the fact that in the ANSCII character set,
   the digits O to 9 have sequential code values. */
void B_FN(printd)(B_TYPE n) {
    B_TYPE a;

    if(n < 0) {
        B_FN(write)('-');
        n = -n;
    }

    if((a = n / 10))
        B_FN(printd)(a);
    B_FN(write)(n % 10 + '0');
}

/* The following function will print an unsigned number, n,
   to the base 8. */
void B_FN(printo)(B_TYPE n) {
    B_TYPE a;

    if((a = (uintptr_t)n >> 3))
        B_FN(printo)(a);
    B_FN(write)((n & 7) + '0');
}

/* The following function is a general formatting, printing, and
   conversion subroutine. The first argument is a format string.
   Character sequences,of the form ‘%x’ are interpreted and cause
   conversion of type x’ of the next argument, other character
   sequences are printed verbatim. */
void B_FN(printf)(B_TYPE fmt, ...) {
    B_TYPE x, c, i = 0, j;

    va_list ap;
    va_start(ap, fmt);
loop:
    while((c = B_FN(_char)(fmt, i++)) != '%') {
        if(c == '\0')
            goto end;
        B_FN(write)(c);
    }
    switch(c = B_FN(_char)(fmt, i++)) {
        case 'd': /* decimal */
            x = va_arg(ap, B_TYPE);
            if(x < 0) {
                x = -x;
                B_FN(write)('-');
            }
            B_FN(printd)(x);
            goto loop;

        case 'o': /* octal */
            x = va_arg(ap, B_TYPE);
            if(x < 0) {
                x = -x;
                B_FN(write)('-');
            }
            B_FN(printo)(x);
            goto loop;

        case 'c':
            x = va_arg(ap, B_TYPE);
            B_FN(write)(x);
            goto loop;

        case 's':
            x = va_arg(ap, B_TYPE);
            j = 0;
            while((c = B_FN(_char)(x, j++)) != '\0')
                B_FN(write)(c);
            goto loop;
        case '%':
            B_FN(write)('%');
            goto loop;
    }
    /* Unknown format. */
    B_FN(write)('%');
    i--;
    goto loop;

end:
    va_end(ap);
}

void B_FN(flush)(void) {
    // Empty.
}
