.globl max
.text
.type max, @function
max:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  sub $8, %rsp
  mov %rdi, -16(%rbp)
  sub $8, %rsp
  mov %rsi, -24(%rbp)
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  cmp %rax, %rdi
  setg %al
  movzb %al, %rax
  cmp $0, %rax
  je .L.else.0
  lea -16(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.max
  jmp .L.end.0
.L.else.0:
  lea -24(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.max
.L.end.0:
  xor %rax, %rax
.L.return.max:
  mov %rbp, %rsp
  pop %rbp
  ret
.globl abs
.text
.type abs, @function
abs:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  sub $8, %rsp
  mov %rdi, -16(%rbp)
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  xor %rax, %rax
  pop %rdi
  cmp %rax, %rdi
  setl %al
  movzb %al, %rax
  cmp $0, %rax
  je .L.else.1
  lea -16(%rbp), %rax
  mov (%rax), %rax
  neg %rax
  jmp .L.return.abs
  jmp .L.end.1
.L.else.1:
.L.end.1:
  lea -16(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.abs
  xor %rax, %rax
.L.return.abs:
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
  sub $8, %rsp
  sub $8, %rsp
  lea -16(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea max(%rip), %rax
  push %rax
  mov $10, %rax
  push %rax
  mov $20, %rax
  push %rax
  pop %rsi
  pop %rdi
  pop %r10
  call *%r10
  pop %rdi
  mov %rax, (%rdi)
  lea -24(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea abs(%rip), %rax
  push %rax
  mov $15, %rax
  neg %rax
  push %rax
  pop %rdi
  pop %r10
  call *%r10
  pop %rdi
  mov %rax, (%rdi)
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  add %rdi, %rax
  jmp .L.return.main
  add $16, %rsp
  xor %rax, %rax
.L.return.main:
  mov %rbp, %rsp
  pop %rbp
  ret
