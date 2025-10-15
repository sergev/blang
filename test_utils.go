package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Prevent "unused" linter warnings in tools that don't analyze test files together.
var _ = []interface{}{
	ensureLibbOrSkip,
	ensureBlangOrSkip,
	writeTempFile,
	createTempBFile,
	compileToLL,
	linkWithClang,
	runExecutable,
	compileLinkRunFromBFile,
	compileLinkRunFromCode,
	runWithTimeout,
	hasSubstring,
	contains,
	min,
}

// ensureLibbOrSkip skips the test if runtime object file is missing.
func ensureLibbOrSkip(t testing.TB) {
	t.Helper()
	if _, err := os.Stat("libb.o"); err != nil {
		t.Skip("libb.o not found, run 'make' first")
	}
}

// ensureBlangOrSkip skips the test if the blang binary is missing.
func ensureBlangOrSkip(t testing.TB) {
	t.Helper()
	if _, err := os.Stat("./blang"); err != nil {
		t.Skip("./blang not found, build the CLI first")
	}
}

// writeTempFile writes a file in the specified directory and returns its path.
func writeTempFile(t testing.TB, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", p, err)
	}
	return p
}

// createTempBFile creates a temporary directory with a B source and returns useful paths.
func createTempBFile(t testing.TB, name, code string) (tmpDir, bFile, llFile, exeFile string) {
	t.Helper()
	tmpDir = t.TempDir()
	bFile = writeTempFile(t, tmpDir, name+".b", code)
	llFile = filepath.Join(tmpDir, name+".ll")
	exeFile = filepath.Join(tmpDir, name)
	return
}

// compileToLL compiles a B source file into an LLVM IR file.
func compileToLL(t testing.TB, input string, llOut string) {
	t.Helper()
	args := NewCompileOptions("blang", []string{input})
	args.OutputFile = llOut
	args.OutputType = OutputIR
	if err := Compile(args); err != nil {
		t.Fatalf("Compile(%s) failed: %v", input, err)
	}
}

// linkWithClang links an LLVM IR file with libb.o into an executable.
func linkWithClang(t testing.TB, llFile, exeFile string) {
	t.Helper()
	cmd := exec.Command("clang", llFile, "libb.o", "-o", exeFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Linking failed: %v\nOutput: %s", err, out)
	}
}

// runExecutable runs an executable and returns its stdout and exit code.
func runExecutable(t testing.TB, exeFile string) ([]byte, int) {
	t.Helper()
	cmd := exec.Command(exeFile)
	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, exitErr.ExitCode()
		}
		t.Fatalf("Failed to run executable: %v", err)
	}
	return stdout, 0
}

// compileLinkRunFromBFile compiles, links, and runs from an existing B file.
func compileLinkRunFromBFile(t testing.TB, bFile string) (string, int) {
	t.Helper()
	ensureLibbOrSkip(t)
	dir := filepath.Dir(bFile)
	llFile := filepath.Join(dir, "test.ll")
	exeFile := filepath.Join(dir, "test")
	compileToLL(t, bFile, llFile)
	linkWithClang(t, llFile, exeFile)
	out, code := runExecutable(t, exeFile)
	return string(out), code
}

// compileLinkRunFromCode compiles, links, and runs from in-memory code, returning stdout.
func compileLinkRunFromCode(t testing.TB, name, code string) string {
	t.Helper()
	ensureLibbOrSkip(t)
	dir, bFile, llFile, exeFile := createTempBFile(t, name, code)
	_ = dir
	compileToLL(t, bFile, llFile)
	linkWithClang(t, llFile, exeFile)
	out, _ := runExecutable(t, exeFile)
	return string(out)
}

// runWithTimeout runs an executable with a timeout and returns its stdout and exit code.
func runWithTimeout(t testing.TB, exeFile string, timeout time.Duration) ([]byte, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, exeFile)
	stdout, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			// Preserve wording from existing tests.
			t.Fatal("Program exceeded 30 second timeout")
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, exitErr.ExitCode()
		}
		t.Fatalf("Failed to run executable: %v", err)
	}
	return stdout, 0
}

// contains reports whether substr is within s.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// hasSubstring is a simple substring search used by tests.
func hasSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---- Lexer test helpers ----

// newTestLexer creates a lexer for the provided input.
func newTestLexer(t testing.TB, input string) *Lexer {
	t.Helper()
	args := NewCompileOptions("test", nil)
	return NewLexer(args, strings.NewReader(input))
}

// lexIdentifier returns the parsed identifier from input.
func lexIdentifier(t testing.TB, input string) string {
	t.Helper()
	l := newTestLexer(t, input)
	got, err := l.Identifier()
	if err != nil {
		t.Fatalf("Identifier() error = %v", err)
	}
	return got
}

// lexNumber returns the parsed number from input.
func lexNumber(t testing.TB, input string) int64 {
	t.Helper()
	l := newTestLexer(t, input)
	got, err := l.Number()
	if err != nil {
		t.Fatalf("Number() error = %v", err)
	}
	return got
}

// lexStringLiteral parses a string literal input, assuming leading '"'.
func lexStringLiteral(t testing.TB, input string) string {
	t.Helper()
	l := newTestLexer(t, input)
	// Skip opening quote
	if _, err := l.ReadChar(); err != nil {
		t.Fatalf("ReadChar() error = %v", err)
	}
	got, err := l.String()
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	return got
}

// lexCharacterLiteral parses a character literal input, assuming leading '\”.
func lexCharacterLiteral(t testing.TB, input string) int64 {
	t.Helper()
	l := newTestLexer(t, input)
	// Skip opening quote
	if _, err := l.ReadChar(); err != nil {
		t.Fatalf("ReadChar() error = %v", err)
	}
	got, err := l.Character()
	if err != nil {
		t.Fatalf("Character() error = %v", err)
	}
	return got
}

// lexWhitespaceNextRune consumes whitespace/comments and returns the next rune.
func lexWhitespaceNextRune(t testing.TB, input string) rune {
	t.Helper()
	l := newTestLexer(t, input)
	if err := l.Whitespace(); err != nil {
		t.Fatalf("Whitespace() error = %v", err)
	}
	got, err := l.ReadChar()
	if err != nil {
		t.Fatalf("ReadChar() error = %v", err)
	}
	return got
}

// lexCommentRest skips an opening comment and returns the remaining text after it.
func lexCommentRest(t testing.TB, input string) string {
	t.Helper()
	l := newTestLexer(t, input)
	// Skip opening /*
	if _, err := l.ReadChar(); err != nil {
		t.Fatalf("ReadChar() error = %v", err)
	}
	if _, err := l.ReadChar(); err != nil {
		t.Fatalf("ReadChar() error = %v", err)
	}
	if err := l.Comment(); err != nil {
		t.Fatalf("Comment() error = %v", err)
	}
	var rest []rune
	for {
		c, err := l.ReadChar()
		if err != nil {
			break
		}
		rest = append(rest, c)
	}
	return string(rest)
}
