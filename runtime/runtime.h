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
word_t b_char(word_t string, word_t i)
	ALIAS("char");
void b_lchar(word_t string, word_t i, word_t chr)
	ALIAS("lchar");
word_t b_read(void)
	ALIAS("read");
word_t b_nread(word_t file, word_t buffer, word_t count)
	ALIAS("nread");
void b_writeb(word_t c)
	ALIAS("writeb");
void b_write(word_t ch)
	ALIAS("write");
word_t b_nwrite(word_t file, word_t buffer, word_t count)
	ALIAS("nwrite");
void b_printd(word_t n)
	ALIAS("printd");
void b_printo(word_t n)
	ALIAS("printo");
void b_printf(word_t fmt, ...)
	ALIAS("printf");
void b_flush(void)
	ALIAS("flush");

//
// Syscall wrapper implementation
//
static inline long syscall(long n, long a1, long a2, long a3)
{
#ifdef __aarch64__
    register long x0 asm("x0") = a1;
    register long x1 asm("x1") = a2;
    register long x2 asm("x2") = a3;
#ifdef linux
    register long x8 asm("x8") = n;          // syscall number goes in x8
    asm volatile("svc #0"                    // svc #0 invokes the kernel
                 : "+r"(x0)                  // x0 is input/output (return)
                 : "r"(x1), "r"(x2), "r"(x8) // inputs
                 : "memory");
#endif
#ifdef __APPLE__
    register long x16 asm("x16") = n;         // syscall number goes in x16
    asm volatile("svc #0x80"                  // svc #0x80 performs the syscall
                 : "+r"(x0)                   // output: x0 returned
                 : "r"(x1), "r"(x2), "r"(x16)
                 : "memory");
#endif
    return x0;
#endif

#ifdef __x86_64__
    long ret;
#ifdef __APPLE__
    n |= 0x2000000;
#endif
    asm volatile("syscall"
                 : "=a"(ret)                         // output: rax <- return
                 : "a"(n), "D"(a1), "S"(a2), "d"(a3) // inputs: rax, rdi, rsi, rdx
                 : "rcx", "r11", "memory"            // syscall clobbers
    );
    return ret;
#endif
}
