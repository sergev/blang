#include "runtime.h"

#ifdef linux
//
// Entry point of any B program.
//
void _start(void) __asm__("_start"); // assure, that _start is really named _start in asm

void _start()
{
    word_t code = main();
    syscall(SYS_exit, code, 0, 0);
}
#endif
