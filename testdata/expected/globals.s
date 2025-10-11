.globl counter
.data
.type counter, @object
.align 8
counter:
  .quad 0
.globl values
.data
.type values, @object
.align 8
values:
  .quad .+8
  .quad 10
  .quad 20
  .quad 30
.globl increment
.text
.type increment, @function
increment:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  lea counter(%rip), %rax
  push %rax
  mov (%rax), %rax
  lea counter(%rip), %rax
  mov (%rax), %rax
  push %rax
  mov $1, %rax
  pop %rdi
  add %rdi, %rax
  pop %rdi
  mov %rax, (%rdi)
  xor %rax, %rax
.L.return.increment:
  mov %rbp, %rsp
  pop %rbp
  ret
.globl sum_values
.text
.type sum_values, @function
sum_values:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  sub $8, %rsp
  sub $8, %rsp
  lea -24(%rbp), %rax
  push %rax
  mov (%rax), %rax
  xor %rax, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -16(%rbp), %rax
  push %rax
  mov (%rax), %rax
  xor %rax, %rax
  pop %rdi
  mov %rax, (%rdi)
.L.start.0:
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  mov $3, %rax
  pop %rdi
  cmp %rax, %rdi
  setl %al
  movzb %al, %rax
  cmp $0, %rax
  je .L.end.0
  lea -24(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea values(%rip), %rax
  push (%rax)
  lea -16(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  shl $3, %rax
  add %rdi, %rax
  mov (%rax), %rax
  pop %rdi
  add %rdi, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -16(%rbp), %rax
  mov (%rax), %rcx
  addq $1, (%rax)
  mov %rcx, %rax
  jmp .L.start.0
.L.end.0:
  lea -24(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.sum_values
  add $16, %rsp
  xor %rax, %rax
.L.return.sum_values:
  mov %rbp, %rsp
  pop %rbp
  ret
.globl main
.text
.type main, @function
main:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  lea increment(%rip), %rax
  push %rax
  pop %r10
  call *%r10
  lea increment(%rip), %rax
  push %rax
  pop %r10
  call *%r10
  lea sum_values(%rip), %rax
  push %rax
  pop %r10
  call *%r10
  jmp .L.return.main
  xor %rax, %rax
.L.return.main:
  mov %rbp, %rsp
  pop %rbp
  ret
