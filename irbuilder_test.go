package main

import (
	"testing"

	"github.com/llir/llvm/ir/constant"
)

func TestGetOrDeclareFunction_WithExtrnReturnsNil(t *testing.T) {
	c := NewCompiler(NewCompileOptions("blang", nil))
	// simulate extrn variable present: a zero-initialized global
	c.DeclareGlobal("foo", constant.NewInt(c.WordType(), 0))
	if fn := c.GetOrDeclareFunction("foo"); fn != nil {
		t.Fatalf("expected nil for function pointer variable, got %v", fn)
	}
}

func TestGetOrDeclareFunction_NewVariadicExternal(t *testing.T) {
	c := NewCompiler(NewCompileOptions("blang", nil))
	fn := c.GetOrDeclareFunction("bar")
	if fn == nil {
		t.Fatalf("expected function to be declared")
	}
	if !fn.Sig.Variadic {
		t.Fatalf("expected variadic signature for auto-declared extern")
	}
	// second call should return cached
	if fn2 := c.GetOrDeclareFunction("bar"); fn2 != fn {
		t.Fatalf("expected cached function on second call")
	}
}

func TestNewBlock_AutoName(t *testing.T) {
	c := NewCompiler(NewCompileOptions("blang", nil))
	fn := c.DeclareFunction("f", nil)
	c.StartFunction(fn)
	defer c.EndFunction()
	b1 := c.NewBlock("")
	b2 := c.NewBlock("")
	if b1 == nil || b2 == nil {
		t.Fatalf("blocks should be non-nil")
	}
	if b1 == b2 {
		t.Fatalf("auto-named blocks must be distinct")
	}
	if b1.Name() == "" || b2.Name() == "" {
		t.Fatalf("auto-named blocks should have non-empty names")
	}
}

func TestDeclareLocalArray_SizeZero(t *testing.T) {
	c := NewCompiler(NewCompileOptions("blang", nil))
	fn := c.DeclareFunction("f", nil)
	c.StartFunction(fn)
	defer c.EndFunction()
	if v := c.DeclareLocalArray("a", 0); v == nil {
		t.Fatal("expected non-nil value for size 0 array")
	}
}
