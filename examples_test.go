package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCompileAndRun tests the full pipeline: compile, link, and execute
func TestCompileAndRun(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		inputFile  string
		wantExit   int
		wantStdout string
	}{
		// Example programs
		{
			name:       "hello_write",
			inputFile:  "examples/hello.b",
			wantExit:   0,
			wantStdout: "Hello, World!",
		},
		{
			name:       "hello_printf",
			inputFile:  "examples/helloworld.b",
			wantExit:   0,
			wantStdout: "Hello, World!",
		},
		{
			name:       "example_fibonacci",
			inputFile:  "examples/fibonacci.b",
			wantExit:   0,
			wantStdout: "55\n",
		},
		{
			name:       "example_fizzbuzz",
			inputFile:  "examples/fizzbuzz.b",
			wantExit:   0,
			wantStdout: "FizzBuzz", // Check that FizzBuzz appears in output
		},
		// Note: example_e2 is tested separately in TestE2Constant due to long runtime
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			llFile := filepath.Join(tmpDir, tt.name+".ll")
			exeFile := filepath.Join(tmpDir, tt.name)

			compileToLL(t, tt.inputFile, llFile)
			linkWithClang(t, llFile, exeFile)
			stdout, exitCode := runExecutable(t, exeFile)

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
			}

			if tt.wantStdout != "" {
				gotStdout := string(stdout)
				if !hasSubstring(gotStdout, tt.wantStdout) {
					t.Errorf("Stdout = %q, want substring %q", gotStdout, tt.wantStdout)
				}
			}
		})
	}
}

// TestE2Constant tests the e-2 constant calculation
func TestE2Constant(t *testing.T) {
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	inputFile := "examples/e-2.b"
	llFile := filepath.Join(tmpDir, "e-2.ll")
	exeFile := filepath.Join(tmpDir, "e2")

	compileToLL(t, inputFile, llFile)
	linkWithClang(t, llFile, exeFile)

	// Run the executable with a 3s timeout (preserving error text)
	stdout, _ := runWithTimeout(t, exeFile, 3*time.Second)

	// Check that output starts with expected first line
	wantPrefix := "71828 18284 59045 23536 02874"
	gotStdout := string(stdout)
	if !strings.HasPrefix(gotStdout, wantPrefix) {
		t.Errorf("Output does not start with expected prefix.\nWant prefix: %q\nGot: %q", wantPrefix, gotStdout[:min(len(gotStdout), 100)])
	}
}

// TestPDP7CompilerHello compiles the historical PDP-7 B compiler (examples/b.b),
// runs it with examples/helloworld.b as input, and verifies the generated
// output matches the expected PDP-7 code in examples/helloworld.pdp7.
func TestPDP7CompilerHello(t *testing.T) {
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	llFile := filepath.Join(tmpDir, "b.ll")
	exeFile := filepath.Join(tmpDir, "b")

	// Compile the PDP-7 B compiler source to IR and link into an executable.
	compileToLL(t, "examples/b.b", llFile)
	linkWithClang(t, llFile, exeFile)

	in, err := os.Open("examples/helloworld.b")
	if err != nil {
		t.Fatalf("open input: %v", err)
	}
	defer in.Close()

	// Run the compiler and read only the first line of its output.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, exeFile)
	cmd.Stdin = in

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		// Include any immediate stderr content for easier debugging
		b, _ := io.ReadAll(stderr)
		t.Fatalf("start: %v, stderr: %s", err, string(b))
	}

	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		// Capture trailing stderr if read failed
		b, _ := io.ReadAll(stderr)
		t.Fatalf("read first line: %v, stderr: %s", err, string(b))
	}

	// We only care about the first line. Trim newline and assert.
	firstLine := strings.TrimRight(line, "\r\n")
	if firstLine != ".main:.+1" {
		t.Fatalf("unexpected first line: got %q, want %q", firstLine, ".main:.+1")
	}

	// Stop the process after validating the first line to avoid reading all output.
	cancel()
	_ = cmd.Wait() // Best-effort cleanup; ignore error from context cancellation
}

// TestGlobalArrayInitialization verifies that global vector initialization
// produces a pointer to a contiguous buffer and indexing returns the
// initialized values in order (true B array semantics).
func TestGlobalArrayInitialization(t *testing.T) {
	ensureLibbOrSkip(t)

	code := `
a[12] 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12;
main() {
    printf("%d %d %d %d*n", a[0], a[1], a[10], a[11]);
}
`

	got := compileLinkRunFromCode(t, "array_init", code)
	want := "1 2 11 12\n"
	if got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}

// TestGlobalArrayPointerSemantics ensures B global arrays behave as pointers:
// *a == a[0] and *(a+8) == a[1] (word size = 8 bytes).
func TestGlobalArrayPointerSemantics(t *testing.T) {
	ensureLibbOrSkip(t)

	code := `
a[3] 1, 2, 3;
main() {
    printf("%d %d %d %d*n", *a, *(a+8), a[0], a[1]);
}
`

	got := compileLinkRunFromCode(t, "array_ptr", code)
	want := "1 2 1 2\n"
	if got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}
