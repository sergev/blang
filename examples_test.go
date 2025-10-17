package main

import (
	"bytes"
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

// The test compiles the historical PDP-7 B compiler (examples/b.b),
// runs it with examples/b.b as input, and verifies the generated
// output matches the expected PDP-7 code in examples/b.pdp7.
func TestPDP7CompilerB(t *testing.T) {
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	llFile := filepath.Join(tmpDir, "b.ll")
	exeFile := filepath.Join(tmpDir, "b")

	// Compile the PDP-7 B compiler source to IR and link into an executable.
	compileToLL(t, "examples/b.b", llFile)
	linkWithClang(t, llFile, exeFile)

	in, err := os.Open("examples/b.b")
	if err != nil {
		t.Fatalf("open input: %v", err)
	}
	defer in.Close()

	cmd := exec.Command(exeFile)
	cmd.Stdin = in
	got, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("program exited with code %d: %s", ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("run: %v", err)
	}

	// Read expected output
	wantPDP7, err := os.ReadFile("examples/b.pdp7")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(got, []byte(wantPDP7)) {
		diffText := buildLineDiff(string(wantPDP7), string(got))
		t.Errorf("output mismatch (-want +got):\n%s", diffText)
	}
}
