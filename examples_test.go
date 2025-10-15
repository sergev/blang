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
	requireLibbO(t)

	tests := []struct {
		Name       string
		InputFile  string
		WantExit   int
		WantStdout string
	}{
		// Example programs
		{
			Name:       "hello_write",
			InputFile:  "examples/hello.b",
			WantExit:   0,
			WantStdout: "Hello, World!",
		},
		{
			Name:       "hello_printf",
			InputFile:  "examples/helloworld.b",
			WantExit:   0,
			WantStdout: "Hello, World!",
		},
		{
			Name:       "example_fibonacci",
			InputFile:  "examples/fibonacci.b",
			WantExit:   0,
			WantStdout: "55\n",
		},
		{
			Name:       "example_fizzbuzz",
			InputFile:  "examples/fizzbuzz.b",
			WantExit:   0,
			WantStdout: "FizzBuzz", // Check that FizzBuzz appears in output
		},
		// Note: example_e2 is tested separately in TestE2Constant due to long runtime
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tmpDir := t.TempDir()
			llFile := filepath.Join(tmpDir, tt.Name+".ll")
			exeFile := filepath.Join(tmpDir, tt.Name)

			// Step 1: Compile B program to LLVM IR
			args := NewCompileOptions("blang", []string{tt.InputFile})
			args.OutputFile = llFile
			args.OutputType = OutputIR

			err := Compile(args)
			if err != nil {
				t.Fatalf("Compile(%s) failed: %v", tt.InputFile, err)
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
			if exitCode != tt.WantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.WantExit)
			}

			// Check stdout if expected
			if tt.WantStdout != "" {
				gotStdout := string(stdout)
				if !hasSubstring(gotStdout, tt.WantStdout) {
					t.Errorf("Stdout = %q, want substring %q", gotStdout, tt.WantStdout)
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
	args.OutputType = OutputIR

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
