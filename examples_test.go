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

// TestCompileAndRun tests the full pipeline: compile, link, and execute
func TestCompileAndRun(t *testing.T) {
	// Check if clang is available
	if _, err := os.Stat("libb.o"); err != nil {
		t.Skip("libb.o not found, run 'make' first")
	}

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

			// Step 1: Compile B program to LLVM IR
			args := NewCompileOptions("blang", []string{tt.inputFile})
			args.OutputFile = llFile

			err := Compile(args)
			if err != nil {
				t.Fatalf("Compile(%s) failed: %v", tt.inputFile, err)
			}

			// Step 2: Link with libb.o using clang
			linkCmd := exec.Command("clang", llFile, "libb.o", "-o", exeFile)
			linkOutput, err := linkCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Linking failed: %v\nOutput: %s", err, linkOutput)
			}

			// Step 3: Run the executable
			runCmd := exec.Command(exeFile)
			stdout, err := runCmd.Output()
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("Failed to run executable: %v", err)
				}
			}

			// Check exit code
			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
			}

			// Check stdout if expected
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
	// Check if clang is available
	if _, err := os.Stat("libb.o"); err != nil {
		t.Skip("libb.o not found, run 'make' first")
	}

	tmpDir := t.TempDir()
	inputFile := "examples/e-2.b"
	llFile := filepath.Join(tmpDir, "e-2.ll")
	exeFile := filepath.Join(tmpDir, "e2")

	// Step 1: Compile B program to LLVM IR
	args := NewCompileOptions("blang", []string{inputFile})
	args.OutputFile = llFile

	err := Compile(args)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Step 2: Link with libb.o using clang
	linkCmd := exec.Command("clang", llFile, "libb.o", "-o", exeFile)
	linkOutput, err := linkCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Linking failed: %v\nOutput: %s", err, linkOutput)
	}

	// Step 3: Run the executable (with 3 second timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	runCmd := exec.CommandContext(ctx, exeFile)
	stdout, err := runCmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatal("Program exceeded 30 second timeout")
		}
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("Failed to run executable: %v", err)
		}
	}

	// Check that output starts with expected first line
	wantPrefix := "71828 18284 59045 23536 02874"
	gotStdout := string(stdout)
	if !strings.HasPrefix(gotStdout, wantPrefix) {
		t.Errorf("Output does not start with expected prefix.\nWant prefix: %q\nGot: %q", wantPrefix, gotStdout[:min(len(gotStdout), 100)])
	}
}
