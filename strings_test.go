package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestStrings tests string and character literal features (from oldtests/string_test.cpp)
func TestStrings(t *testing.T) {
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
			name: "string_literals",
			code: `sa "*t*0x";
			sb "foo*ebar";

			main() {
				extrn sa, sb;

				printf("*(*)***"*n");
				printf("%d %d %d*n", char(sa, 0), char(sa, 1), char(sa, 2));
				printf("%d %d %d %d %d*n", char(sb, 0), char(sb, 1), char(sb, 2), char(sb, 3), char(sb, 4));
			}`,
			wantStdout: `{}*"
9 0 120
102 111 111 0 98
`,
		},
		{
			name: "char_literals",
			code: `main() {
				printf("%d*n", '*0');
				printf("%d*n", '*e');
				printf("%d*n", '*t');
				printf("%d*n", '*n');
				printf("%d*n", '*r');
				printf("%c*n", '*(*)***'*"');
			}`,
			wantStdout: `0
0
9
10
13
{}*'"
`,
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
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}
