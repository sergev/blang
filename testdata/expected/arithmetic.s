.globl add
.text
.type add, @function
add:
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
  add %rdi, %rax
  jmp .L.return.add
  xor %rax, %rax
.L.return.add:
  mov %rbp, %rsp
  pop %rbp
  ret
.globl sub
.text
.type sub, @function
sub:
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
  mov %rax, %rdi
  pop %rax
  sub %rdi, %rax
  jmp .L.return.sub
  xor %rax, %rax
.L.return.sub:
  mov %rbp, %rsp
  pop %rbp
  ret
.globl mul
.text
.type mul, @function
mul:
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
  imul %rdi, %rax
  jmp .L.return.mul
  xor %rax, %rax
.L.return.mul:
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
  sub $8, %rsp
  sub $8, %rsp
  lea -16(%rbp), %rax
  push %rax
  mov (%rax), %rax
  mov $10, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -24(%rbp), %rax
  push %rax
  mov (%rax), %rax
  mov $20, %rax
  pop %rdi
  mov %rax, (%rdi)
  lea -32(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea add(%rip), %rax
  push %rax
  lea -16(%rbp), %rax
  mov (%rax), %rax
  push %rax
  lea -24(%rbp), %rax
  mov (%rax), %rax
  push %rax
  pop %rsi
  pop %rdi
  pop %r10
  call *%r10
  pop %rdi
  mov %rax, (%rdi)
  lea -32(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea sub(%rip), %rax
  push %rax
  lea -32(%rbp), %rax
  mov (%rax), %rax
  push %rax
  mov $5, %rax
  push %rax
  pop %rsi
  pop %rdi
  pop %r10
  call *%r10
  pop %rdi
  mov %rax, (%rdi)
  lea -32(%rbp), %rax
  push %rax
  mov (%rax), %rax
  lea mul(%rip), %rax
  push %rax
  lea -32(%rbp), %rax
  mov (%rax), %rax
  push %rax
  mov $2, %rax
  push %rax
  pop %rsi
  pop %rdi
  pop %r10
  call *%r10
  pop %rdi
  mov %rax, (%rdi)
  lea -32(%rbp), %rax
  mov (%rax), %rax
  jmp .L.return.main
  add $32, %rsp
  xor %rax, %rax
.L.return.main:
  mov %rbp, %rsp
  pop %rbp
  ret
