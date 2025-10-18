//
// Syscall wrapper implementation for ARM64
//
static inline long syscall(long n, long a1, long a2, long a3)
{
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
}
