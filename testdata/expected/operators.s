.globl test_bitwise
.text
.type test_bitwise, @function
test_bitwise:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  sub $8, %rsp
  mov %rdi, -16(%rbp)
  sub $8, %rsp
  mov %rsi, -24(%rbp)
  sub $8, %rsp
  sub $8, %rsp
  sub $8, %rsp
  sub $8, %rsp
  lea -32(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  and %rdi, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -40(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  or %rdi, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -48(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  mov $2, %rax
  mov %rax, %rcx
  pop %rax
  shl %cl, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -32(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -40(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  add %rdi, %rax
  push %rax
  lea -48(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  add %rdi, %rax
  jmp .L.return.test_bitwise
  add $32, %rsp
  xor %rax, %rax
.L.return.test_bitwise:
  mov %rbp, %rsp
  pop %rbp
  ret
.globl test_comparison
.text
.type test_comparison, @function
test_comparison:
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
  sete %al
  movzb %al, %rax
  cmp $0, %rax
  je .L.else.0
  xor %rax, %rax
  jmp .L.return.test_comparison
  jmp .L.end.0
.L.else.0:
.L.end.0:
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  pop %rdi
  cmp %rax, %rdi
  setne %al
  movzb %al, %rax
  cmp $0, %rax
  je .L.else.1
  mov $1, %rax
  jmp .L.return.test_comparison
  jmp .L.end.1
.L.else.1:
.L.end.1:
  mov $2, %rax
  jmp .L.return.test_comparison
  xor %rax, %rax
.L.return.test_comparison:
  mov %rbp, %rsp
  pop %rbp
  ret
.globl test_unary
.text
.type test_unary, @function
test_unary:
  push %rbp
  mov %rsp, %rbp
  sub $8, %rsp
  sub $8, %rsp
  sub $8, %rsp
  lea -16(%rbp), %rax
  push %rax
  mov (%rax), %rax
  mov $10, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -16(%rbp), %rax
  mov (%rax), %rcx
  addq $1, (%rax)
  mov %rcx, %rax
  lea -16(%rbp), %rax
  mov (%rax), %rdi
  add $1, %rdi
  mov %rdi, (%rax)
  mov (%rax), %rax
  lea -16(%rbp), %rax
  mov (%rax), %rcx
  subq $1, (%rax)
  mov %rcx, %rax
  lea -16(%rbp), %rax
  mov (%rax), %rdi
  sub $1, %rdi
  mov %rdi, (%rax)
  mov (%rax), %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.test_unary
  add $16, %rsp
  xor %rax, %rax
.L.return.test_unary:
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
  lea test_bitwise(%rip), %rax
  push %rax
  mov $12, %rax
  push %rax
  mov $5, %rax
  push %rax
  pop %rsi
  pop %rdi
  pop %r10
  call *%r10
  pop %rdi
  mov %rax, (%rdi)
  lea -16(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea test_comparison(%rip), %rax
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
  add %rdi, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -16(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea test_unary(%rip), %rax
  push %rax
  pop %r10
  call *%r10
  pop %rdi
  add %rdi, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -16(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.main
  add $16, %rsp
  xor %rax, %rax
.L.return.main:
  mov %rbp, %rsp
  pop %rbp
  ret
