//
// Internal details of B standard library
//
#include <stdarg.h>
#include <stdint.h>
#include <sys/syscall.h>

//
// Set external name of the symbol
//
#ifdef linux
#define ALIAS(name) __asm__("b."name)
#endif
#ifdef __APPLE__
#define ALIAS(name) __asm__("_b."name)
#endif

//
// Type representing B's word-sized value.
//
typedef intptr_t word_t;

// Select output stream: 0-stdout, 1-stderr.
extern word_t b_fout
    ALIAS("fout");

//
// Function declarations.
//
void b_exit(void)
    ALIAS("exit");
word_t b_char(word_t string, /*word_t i,*/ ...)
    ALIAS("char");
void b_lchar(word_t string, /*word_t i, word_t chr,*/ ...)
    ALIAS("lchar");
word_t b_read(void)
    ALIAS("read");
word_t b_nread(word_t file, /*word_t buffer, word_t count,*/ ...)
    ALIAS("nread");
void b_writeb(word_t c, ...)
    ALIAS("writeb");
void b_write(word_t ch, ...)
    ALIAS("write");
word_t b_nwrite(word_t file, /*word_t buffer, word_t count,*/ ...)
    ALIAS("nwrite");
void b_printd(word_t n, ...)
    ALIAS("printd");
void b_printo(word_t n, ...)
    ALIAS("printo");
void b_printf(word_t fmt, ...)
    ALIAS("printf");
void b_flush(void)
    ALIAS("flush");

//
// Inline functions.
//
#ifdef __x86_64__
#include "x86_64.h"
#endif
#ifdef __aarch64__
#include "aarch64.h"
#endif
