.globl factorial
.text
.type factorial, @function
factorial:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  sub $8, %rsp
  mov %rdi, -16(%rbp)
  sub $8, %rsp
  sub $8, %rsp
  sub $8, %rsp
  lea -24(%rbp), %rax
  push %rax
  mov (%rax), %rax
  mov $1, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -32(%rbp), %rax
  push %rax
  mov (%rax), %rax
  mov $1, %rax
  pop %rdi
  mov %rax, (%rdi)
.L.start.0:
  lea -32(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  cmp %rax, %rdi
  setle %al
  movzb %al, %rax
  cmp $0, %rax
  je .L.end.0
  lea -24(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -32(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  imul %rdi, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -32(%rbp), %rax
  mov (%rax), %rcx
  addq $1, (%rax)
  mov %rcx, %rax
  jmp .L.start.0
.L.end.0:
  lea -24(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.factorial
  add $24, %rsp
  xor %rax, %rax
.L.return.factorial:
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
  lea factorial(%rip), %rax
  push %rax
  mov $5, %rax
  push %rax
  pop %rdi
  pop %r10
  call *%r10
  jmp .L.return.main
  xor %rax, %rax
.L.return.main:
  mov %rbp, %rsp
  pop %rbp
  ret
