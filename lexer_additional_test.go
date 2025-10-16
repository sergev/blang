package main

import (
	"strings"
	"testing"
)

func TestLexer_ExpectChar_Branches(t *testing.T) {
	l := newTestLexer(t, ")")
	if err := l.ExpectChar(')', "expect ')'"); err != nil {
		t.Fatalf("want success, got %v", err)
	}

	l = newTestLexer(t, "x")
	if err := l.ExpectChar(')', "oops"); err == nil || !strings.Contains(err.Error(), "oops, got '") {
		t.Fatalf("want mismatch error, got %v", err)
	}

	l = newTestLexer(t, "")
	if err := l.ExpectChar(')', "EOF msg"); err == nil || !strings.Contains(err.Error(), "EOF msg") {
		t.Fatalf("want EOF error, got %v", err)
	}
}
