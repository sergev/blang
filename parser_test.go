package main

import (
	"strings"
	"testing"
)

// helpers to drive parser directly
func parseOK(t *testing.T, src string) error {
	t.Helper()
	args := NewCompileOptions("blang", nil)
	c := NewCompiler(args)
	l := NewLexer(args, strings.NewReader(src))
	return ParseDeclarations(l, c)
}

func parseErr(t *testing.T, src, want string) error {
	t.Helper()
	args := NewCompileOptions("blang", nil)
	c := NewCompiler(args)
	l := NewLexer(args, strings.NewReader(src))
	err := ParseDeclarations(l, c)
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if want != "" && !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %q", want, err.Error())
	}
	return err
}

func TestParseReturn_VariantsAndErrors(t *testing.T) {
	if err := parseOK(t, `f(){ return; } g(){ return(42); }`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parseErr(t, `f(){ return(42; }`, "expect ')'")
	parseErr(t, `f(){ return(42) }`, "expect ';'")
}

func TestParseArguments_ListAndErrors(t *testing.T) {
	if err := parseOK(t, `f(a,b,c) { return(0); }`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parseErr(t, `f(a b) {}`, "unexpected character")
	// To trigger parseArguments error, provide a comma with no identifier
	parseErr(t, `f(,) {}`, "expect ')' or identifier")
}

func TestParseIvalConst_AllKinds(t *testing.T) {
	// identifier, char, string, negative number, number
	src := `
g  /*TODO id,*/ 'x', "str", -123, 456;
id 0;
`
	if err := parseOK(t, src); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// lone '-' at EOF should error
	parseErr(t, `g -`, "expect ival")
}

func TestParseVector_SizeAndDefaults(t *testing.T) {
	if err := parseOK(t, `v[3] 1,2;`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := parseOK(t, `w[] 1,2,3;`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parseErr(t, `q[3 1,2;`, "expect ']' after vector size")
}

func TestParseExtrn_ListAndErrors(t *testing.T) {
	if err := parseOK(t, `main(){ extrn a,b; }`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parseErr(t, `main(){ extrn a:; }`, "expect ';' or ','")
}

func TestParseCase_NumberAndChar(t *testing.T) {
	ok := `
f(x){
    switch(x){
        case 1:;
        case 'a':;
    }
}
`
	if err := parseOK(t, ok); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 'case' outside switch
	parseErr(t, `main(){ case 1:; }`, "case' outside of 'switch")
}
