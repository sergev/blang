package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNestedLoops tests nested while loops with unique labels
func TestNestedLoops(t *testing.T) {
	// Check if clang is available
	if _, err := os.Stat("libb.o"); err != nil {
		t.Skip("libb.o not found, run 'make' first")
	}

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "nested_while_basic",
			code: `main() {
				auto i, j, sum;

				sum = 0;
				i = 0;
				while (i < 3) {
					j = 0;
					while (j < 3) {
						sum = sum + 1;
						j++;
					}
					i++;
				}

				printf("sum = %d*n", sum);
			}`,
			wantStdout: "sum = 9\n",
		},
		{
			name: "nested_while_complex",
			code: `main() {
				auto i, j, k, count;

				count = 0;
				i = 1;
				while (i <= 2) {
					j = 1;
					while (j <= 2) {
						k = 1;
						while (k <= 2) {
							count++;
							k++;
						}
						j++;
					}
					i++;
				}

				printf("count = %d*n", count);
			}`,
			wantStdout: "count = 8\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test.b")
			llFile := filepath.Join(tmpDir, "test.ll")
			exeFile := filepath.Join(tmpDir, "test")

			// Write test code to file
			err := os.WriteFile(inputFile, []byte(tt.code), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Step 1: Compile B program to LLVM IR
			args := NewCompileOptions("blang", []string{inputFile})
			args.OutputFile = llFile

			err = Compile(args)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
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
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					t.Fatalf("Failed to run executable: %v", err)
				}
			}

			// Check stdout
			gotStdout := string(stdout)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}
