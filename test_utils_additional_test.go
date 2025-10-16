package main

import (
	"strings"
	"testing"
)

func TestMin(t *testing.T) {
	if min(1, 2) != 1 {
		t.Fatal("min(1,2) != 1")
	}
	if min(3, -1) != -1 {
		t.Fatal("min(3,-1) != -1")
	}
}

func TestCompileLinkRunFromBFile_Smoke(t *testing.T) {
	ensureLibbOrSkip(t)
	tmp := t.TempDir()
	b := writeTempFile(t, tmp, "x.b", "main(){ write('x'); return(0); }")
	out, code := compileLinkRunFromBFile(t, b)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d", code)
	}
	if !strings.Contains(out, "x") {
		t.Fatalf("stdout does not contain expected output, got %q", out)
	}
}
