#include "runtime.h"

//
// The current process is terminated.
//
void exit()
{
    syscall(SYS_exit, 0, 0, 0);
}
