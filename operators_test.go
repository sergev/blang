package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCompoundAssignments tests compound assignment operators (from oldtests/assignment_test.cpp)
func TestCompoundAssignments(t *testing.T) {

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
			name: "assign_add",
			code: `main() {
				auto x;
				x = 10;
				x =+ 5;
				printf("%d*n", x);
			}`,
			wantStdout: "15\n",
		},
		{
			name: "assign_subtract",
			code: `main() {
				auto x;
				x = 10;
				x =- 3;
				printf("%d*n", x);
			}`,
			wantStdout: "7\n",
		},
		{
			name: "assign_multiply",
			code: `main() {
				auto x;
				x = 4;
				x =* 3;
				printf("%d*n", x);
			}`,
			wantStdout: "12\n",
		},
		{
			name: "assign_divide",
			code: `main() {
				auto x;
				x = 15;
				x =/ 3;
				printf("%d*n", x);
			}`,
			wantStdout: "5\n",
		},
		{
			name: "assign_modulo",
			code: `main() {
				auto x;
				x = 17;
				x =% 5;
				printf("%d*n", x);
			}`,
			wantStdout: "2\n",
		},
		{
			name: "assign_shift_left",
			code: `main() {
				auto x;
				x = 2;
				x =<< 2;
				printf("%d*n", x);
			}`,
			wantStdout: "8\n",
		},
		{
			name: "assign_shift_right",
			code: `main() {
				auto x;
				x = 16;
				x =>> 2;
				printf("%d*n", x);
			}`,
			wantStdout: "4\n",
		},
		{
			name: "assign_less_or_equal",
			code: `main() {
				auto x;
				x = 5;
				x =<= 3;
				printf("%d*n", x);
				x = 2;
				x =<= 3;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_less_than",
			code: `main() {
				auto x;
				x = 5;
				x =< 5;
				printf("%d*n", x);
				x = 4;
				x =< 5;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_greater_or_equal",
			code: `main() {
				auto x;
				x = 3;
				x =>= 5;
				printf("%d*n", x);
				x = 5;
				x =>= 5;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_greater_than",
			code: `main() {
				auto x;
				x = 4;
				x => 5;
				printf("%d*n", x);
				x = 6;
				x => 5;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_not_equal",
			code: `main() {
				auto x;
				x = 5;
				x =!= 5;
				printf("%d*n", x);
				x = 5;
				x =!= 3;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_equal",
			code: `main() {
				auto x;
				x = 5;
				x === 5;
				printf("%d*n", x);
				x = 5;
				x === 6;
				printf("%d*n", x);
			}`,
			wantStdout: "1\n0\n",
		},
		{
			name: "assign_and",
			code: `main() {
				auto x;
				x = 12;
				x =& 10;
				printf("%d*n", x);
			}`,
			wantStdout: "8\n",
		},
		{
			name: "assign_or",
			code: `main() {
				auto x;
				x = 12;
				x =| 10;
				printf("%d*n", x);
			}`,
			wantStdout: "14\n",
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
			args.OutputType = OutputIR

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
				t.Errorf("Stdout = %q, want %q", gotStdout, tt.wantStdout)
			}
		})
	}
}
