package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGlobals tests global and local variable features (from oldtests/globals_test.cpp)
func TestGlobals(t *testing.T) {
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
			name: "global_scalars",
			code: `a;
			b 123;
			c -345, 'foo', "bar";

			main() {
				extrn a, b, c;

				printf("a = %d*n", a);
				printf("b = %d*n", b);
				a = &c;
				printf("c = %d, '%c', *"%s*"*n", c, a[1], a[2]);
			}`,
			wantStdout: `a = 0
b = 123
c = -345, 'foo', "bar"
`,
		},
		{
			name: "global_vectors",
			code: `a[];
			b[] 123;
			c[4] -345, 'foo', "bar";

			main() {
				extrn a, b, c;

				printf("a = %d*n", a[1]);
				printf("b = %d*n", b[0]);
				printf("c = %d, '%c', *"%s*", %d*n", c[0], c[1], c[2], c[3]);
			}`,
			wantStdout: `a = 123
b = 123
c = -345, 'foo', "bar", 0
`,
		},
		{
			name: "local_scalars",
			code: `main() {
				auto a;
				auto b;
				auto c;

				printf("offset a = %d*n", (&a) - &c);
				printf("offset b = %d*n", (&b) - &c);
				printf("offset c = %d*n", (&c) - &c);
			}`,
			wantStdout: `offset a = 16
offset b = 8
offset c = 0
`,
		},
		{
			name: "local_vectors",
			code: `main() {
				auto a[124];
				auto b['x'];
				auto c[1];

				printf("offset a = %d*n", (&a) - &c);
				printf("offset b = %d*n", (&b) - &c);
				printf("offset c = %d*n", (&c) - &c);
			}`,
			wantStdout: `offset a = 984
offset b = 16
offset c = 0
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
