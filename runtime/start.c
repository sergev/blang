#include "runtime.h"

#ifdef linux
//
// Entry point of any B program.
//
void b_start(void) __asm__("_start"); // assure, that _start is really named _start in asm

void b_start()
{
	word_t main(void);

    word_t code = main();
    syscall(SYS_exit, code, 0, 0);
}
#endif
