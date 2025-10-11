.globl messages
.data
.type messages, @object
.align 8
messages:
  .quad .+8
  .quad .string.0
  .quad .string.1
  .quad .string.2
.globl main
.text
.type main, @function
main:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  lea write(%rip), %rax
  push %rax
  lea messages(%rip), %rax
  push (%rax)
  xor %rax, %rax
  pop %rdi
  shl $3, %rax
  add %rdi, %rax
  mov (%rax), %rax
  push %rax
  pop %rdi
  pop %r10
  call *%r10
  lea write(%rip), %rax
  push %rax
  lea messages(%rip), %rax
  push (%rax)
  mov $1, %rax
  pop %rdi
  shl $3, %rax
  add %rdi, %rax
  mov (%rax), %rax
  push %rax
  pop %rdi
  pop %r10
  call *%r10
  lea write(%rip), %rax
  push %rax
  lea messages(%rip), %rax
  push (%rax)
  mov $2, %rax
  pop %rdi
  shl $3, %rax
  add %rdi, %rax
  mov (%rax), %rax
  push %rax
  pop %rdi
  pop %r10
  call *%r10
  xor %rax, %rax
.L.return.main:
  mov %rbp, %rsp
  pop %rbp
  ret
.section .rodata
.string.0:
  .byte 72
  .byte 101
  .byte 108
  .byte 108
  .byte 111
  .byte 0
.string.1:
  .byte 87
  .byte 111
  .byte 114
  .byte 108
  .byte 100
  .byte 0
.string.2:
  .byte 84
  .byte 101
  .byte 115
  .byte 116
  .byte 0
