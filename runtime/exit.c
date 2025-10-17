#include "runtime.h"

//
// The current process is terminated.
//
void b_exit()
{
    syscall(SYS_exit, 0, 0, 0);
}
