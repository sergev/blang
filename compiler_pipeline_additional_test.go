package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileToIR_Errors(t *testing.T) {
	tmp := t.TempDir()

	// Non-.b input should fail in IR mode
	ll := writeTempFile(t, tmp, "x.ll", "define i64 @main(){ ret i64 0 }")
	args := NewCompileOptions("blang", []string{ll})
	args.OutputType = OutputIR
	args.OutputFile = filepath.Join(tmp, "x.ll.out")
	if err := compileToIR(args); err == nil {
		t.Fatal("expected error for non-.b input in IR mode")
	}

	// Multiple inputs with -o should fail in IR mode
	b1 := writeTempFile(t, tmp, "a.b", "main(){return(0);}")
	b2 := writeTempFile(t, tmp, "b.b", "main(){return(0);}")
	args = NewCompileOptions("blang", []string{b1, b2})
	args.OutputType = OutputIR
	args.OutputFile = filepath.Join(tmp, "out.ll")
	if err := compileToIR(args); err == nil || !strings.Contains(err.Error(), "multiple input files") {
		t.Fatalf("expected multi-input error, got %v", err)
	}
}

func TestCompileToAssembly_SuccessAndErrors(t *testing.T) {
	tmp := t.TempDir()

	// .b -> .s
	b := writeTempFile(t, tmp, "x.b", "main(){return(0);}")
	outS := filepath.Join(tmp, "x.s")
	args := NewCompileOptions("blang", []string{b})
	args.OutputType = OutputAssembly
	args.OutputFile = outS
	if err := compileToAssembly(args); err != nil {
		t.Fatalf("compileToAssembly(.b) err: %v", err)
	}

	// .ll -> .s
	ll := writeTempFile(t, tmp, "y.ll", "define i64 @main(){ ret i64 0 }")
	args = NewCompileOptions("blang", []string{ll})
	args.OutputType = OutputAssembly
	args.OutputFile = filepath.Join(tmp, "y.s")
	if err := compileToAssembly(args); err != nil {
		t.Fatalf("compileToAssembly(.ll) err: %v", err)
	}

	// invalid extension error
	bad := writeTempFile(t, tmp, "z.txt", "nope")
	args = NewCompileOptions("blang", []string{bad})
	args.OutputType = OutputAssembly
	args.OutputFile = filepath.Join(tmp, "z.s")
	if err := compileToAssembly(args); err == nil || !strings.Contains(err.Error(), "must have .b or .ll") {
		t.Fatalf("expected extension error, got %v", err)
	}
}

func TestCompileToObject_SuccessAndErrors(t *testing.T) {
	tmp := t.TempDir()

	// .b -> .o
	b := writeTempFile(t, tmp, "x.b", "main(){return(0);}")
	outO := filepath.Join(tmp, "x.o")
	args := NewCompileOptions("blang", []string{b})
	args.OutputType = OutputObject
	args.OutputFile = outO
	if err := compileToObject(args); err != nil {
		t.Fatalf("compileToObject(.b) err: %v", err)
	}

	// .ll -> .o
	ll := writeTempFile(t, tmp, "y.ll", "define i64 @main(){ ret i64 0 }")
	args = NewCompileOptions("blang", []string{ll})
	args.OutputType = OutputObject
	args.OutputFile = filepath.Join(tmp, "y.o")
	if err := compileToObject(args); err != nil {
		t.Fatalf("compileToObject(.ll) err: %v", err)
	}

	// .s -> .o (minimal asm)
	// Use AT&T syntax for clang on macOS; define a tiny _main returning 0
	s := writeTempFile(t, tmp, "z.s", ".globl _main\n_main:\n  xorl %eax, %eax\n  ret")
	args = NewCompileOptions("blang", []string{s})
	args.OutputType = OutputObject
	args.OutputFile = filepath.Join(tmp, "z.o")
	if err := compileToObject(args); err != nil {
		t.Fatalf("compileToObject(.s) err: %v", err)
	}

	// invalid extension error
	bad := writeTempFile(t, tmp, "q.dat", "???")
	args = NewCompileOptions("blang", []string{bad})
	args.OutputType = OutputObject
	args.OutputFile = filepath.Join(tmp, "q.o")
	if err := compileToObject(args); err == nil || !strings.Contains(err.Error(), "must have .b, .ll or .s") {
		t.Fatalf("expected extension error, got %v", err)
	}
}

func TestCompileToExecutable_UnsupportedExtension(t *testing.T) {
	tmp := t.TempDir()
	bad := writeTempFile(t, tmp, "x.txt", "nope")
	args := NewCompileOptions("blang", []string{bad})
	args.OutputType = OutputExecutable
	args.OutputFile = filepath.Join(tmp, "x")
	if err := compileToExecutable(args); err == nil || !strings.Contains(err.Error(), "unsupported input file extension") {
		t.Fatalf("expected ext error, got %v", err)
	}
}
