//
// Syscall wrapper implementation for RISC-V 64-bit
//
static inline long syscall(long n, long arg1, long arg2, long arg3)
{
    register long a0 asm("a0") = arg1;
    register long a1 asm("a1") = arg2;
    register long a2 asm("a2") = arg3;
    register long a7 asm("a7") = n;  // syscall number goes in a7

    asm volatile ("ecall"                     // ecall invokes the kernel
                  : "+r"(a0)                  // a0 is input/output (return value)
                  : "r"(a1), "r"(a2), "r"(a7) // inputs placed in their registers
                  : "memory");
    return a0;
}
