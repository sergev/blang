.globl main
.text
.type main, @function
main:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  lea write(%rip), %rax
  push %rax
  lea .string.0(%rip), %rax
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
  .byte 44
  .byte 32
  .byte 87
  .byte 111
  .byte 114
  .byte 108
  .byte 100
  .byte 33
  .byte 10
  .byte 0
