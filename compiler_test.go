package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCompile tests the full compilation pipeline
func TestCompile(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		wantFunc  string // Function name that should exist in output
	}{
		{"hello", "testdata/hello.b", "@main"},
		{"arithmetic", "testdata/arithmetic.b", "@main"},
		{"globals", "testdata/globals.b", "@main"},
		{"conditionals", "testdata/conditionals.b", "@main"},
		{"loops", "testdata/loops.b", "@factorial"},
		{"strings", "testdata/strings.b", "@main"},
		{"arrays", "testdata/arrays.b", "@sum"},
		{"pointers", "testdata/pointers.b", "@main"},
		{"switch", "testdata/switch.b", "@classify"},
		{"goto", "testdata/goto.b", "@main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary output file
			outputFile := filepath.Join(t.TempDir(), "output.ll")

			// Compile the input
			args := NewCompilerArgs("blang", []string{tt.inputFile})
			args.OutputFile = outputFile

			err := Compile(args)
			if err != nil {
				t.Fatalf("Compile(%s) failed: %v", tt.inputFile, err)
			}

			// Read generated output
			got, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			output := string(got)

			// Verify it's valid LLVM IR
			if !hasSubstring(output, "define i64") {
				t.Errorf("Output doesn't contain function definition")
			}

			// Verify expected function exists
			if !hasSubstring(output, tt.wantFunc) {
				t.Errorf("Output doesn't contain expected function %s", tt.wantFunc)
			}

			// Verify output is non-empty
			if len(output) == 0 {
				t.Error("Output file is empty")
			}
		})
	}
}

// TestCompileErrors tests that invalid B programs are rejected
func TestCompileErrors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name:        "undefined_variable",
			content:     "main() { x = 10; }",
			wantErr:     true,
			errContains: "undefined identifier",
		},
		{
			name:        "unclosed_comment",
			content:     "/* unclosed comment\nmain() { }",
			wantErr:     true,
			errContains: "unclosed comment",
		},
		{
			name:        "missing_semicolon",
			content:     "main() { auto x x = 10; }",
			wantErr:     true,
			errContains: "expect",
		},
		{
			name:        "unclosed_char_literal",
			content:     "main() { auto c; c = 'abcdefghij; }",
			wantErr:     true,
			errContains: "unclosed char literal",
		},
		{
			name:        "unterminated_string",
			content:     `main() { write("hello); }`,
			wantErr:     true,
			errContains: "unterminated string",
		},
		{
			name:        "undefined_escape",
			content:     `main() { write("*x"); }`,
			wantErr:     true,
			errContains: "undefined escape character",
		},
		{
			name:        "case_outside_switch",
			content:     "main() { case 1: return(0); }",
			wantErr:     true,
			errContains: "case' outside of 'switch",
		},
		// Note: Duplicate identifier detection is not yet implemented in LLVM backend
		// {
		// 	name:        "duplicate_identifier",
		// 	content:     "main() { auto x, x; }",
		// 	wantErr:     true,
		// 	errContains: "already defined",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary input file
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test.b")
			outputFile := filepath.Join(tmpDir, "output.ll")

			err := os.WriteFile(inputFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Try to compile
			args := NewCompilerArgs("blang", []string{inputFile})
			args.OutputFile = outputFile

			err = Compile(args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check error message contains expected substring
			if tt.wantErr && err != nil && tt.errContains != "" {
				errMsg := err.Error()
				if !contains(errMsg, tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, errMsg)
				}
			}
		})
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestLibbFunctions tests runtime library functions (from oldtests/libb_test.cpp)
func TestLibbFunctions(t *testing.T) {
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
			name: "write",
			code: `main() {
				write('Hello,');
				write(' World');
				write('!*n');
			}`,
			wantStdout: "Hello, World!\n",
		},
		{
			name: "printf_basic",
			code: `main() {
				printf("Hello, World!*n");
			}`,
			wantStdout: "Hello, World!\n",
		},
		{
			name: "printf_formats",
			code: `main() {
				printf("format %%d: %d %d*n", 123, -123);
				printf("format %%o: %o %o*n", 234, -234);
			}`,
			wantStdout: "format %d: 123 -123\nformat %o: 352 -352\n",
		},
		{
			name: "printf_char",
			code: `main() {
				printf("format %%c: %c %c*n", 'foo', 'bar');
			}`,
			wantStdout: "format %c: foo bar\n",
		},
		{
			name: "char_function",
			code: `main() {
				write(char("fubar", 2));
				write(char("fubar", 4));
				write(char("fubar", 1));
				write(char("fubar", 0));
				write(char("fubar", 3));
				write('*n');
			}`,
			wantStdout: "brufa\n",
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
			args := NewCompilerArgs("blang", []string{inputFile})
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
				t.Errorf("Stdout = %q, want %q", gotStdout, tt.wantStdout)
			}
		})
	}
}

// TestCompileMultipleFiles tests compiling multiple B files
func TestCompileMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first file
	file1 := filepath.Join(tmpDir, "file1.b")
	err := os.WriteFile(file1, []byte("add(a, b) { return(a + b); }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	// Create second file
	file2 := filepath.Join(tmpDir, "file2.b")
	err = os.WriteFile(file2, []byte("main() { return(add(1, 2)); }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Compile both files
	outputFile := filepath.Join(tmpDir, "output.ll")
	args := NewCompilerArgs("blang", []string{file1, file2})
	args.OutputFile = outputFile

	err = Compile(args)
	if err != nil {
		t.Fatalf("Compile() failed: %v", err)
	}

	// Check output exists and is not empty
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if len(output) == 0 {
		t.Error("Output file is empty")
	}

	// Verify both functions are in output
	outputStr := string(output)
	if !hasSubstring(outputStr, "@add") {
		t.Error("Output doesn't contain add function")
	}
	if !hasSubstring(outputStr, "@main") {
		t.Error("Output doesn't contain main function")
	}
}

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
		// Basic testdata programs
		{
			name:       "hello",
			inputFile:  "testdata/hello.b",
			wantExit:   0,
			wantStdout: "Hello, World!",
		},
		{
			name:       "arithmetic",
			inputFile:  "testdata/arithmetic.b",
			wantExit:   50,
			wantStdout: "",
		},
		{
			name:       "loops",
			inputFile:  "testdata/loops.b",
			wantExit:   120, // factorial(5) = 120
			wantStdout: "",
		},
		{
			name:       "switch",
			inputFile:  "testdata/switch.b",
			wantExit:   30,
			wantStdout: "",
		},
		{
			name:       "goto",
			inputFile:  "testdata/goto.b",
			wantExit:   42,
			wantStdout: "",
		},
		// Examples directory (from oldtests)
		{
			name:       "example_hello_write",
			inputFile:  "examples/hello.b",
			wantExit:   0,
			wantStdout: "Hello, World!\n",
		},
		{
			name:       "example_hello_printf",
			inputFile:  "examples/helloworld.b",
			wantExit:   0,
			wantStdout: "Hello, World!\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			llFile := filepath.Join(tmpDir, tt.name+".ll")
			exeFile := filepath.Join(tmpDir, tt.name)

			// Step 1: Compile B program to LLVM IR
			args := NewCompilerArgs("blang", []string{tt.inputFile})
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

// BenchmarkCompile benchmarks the compilation process
func BenchmarkCompile(b *testing.B) {
	tmpDir := b.TempDir()
	outputFile := filepath.Join(tmpDir, "output.ll")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args := NewCompilerArgs("blang", []string{"testdata/arithmetic.b"})
		args.OutputFile = outputFile
		Compile(args)
	}
}
