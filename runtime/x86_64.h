//
// Syscall wrapper implementation for x86_64 (Intel/AMD)
//
static inline long syscall(long n, long a1, long a2, long a3)
{
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
}
