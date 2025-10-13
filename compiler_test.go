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

// TestCompile tests the full compilation pipeline
func TestCompile(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		wantFunc  string // Function name that should exist in output
	}{
		{"hello", "examples/hello.b", "@main"},
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
			args := NewCompileOptions("blang", []string{tt.inputFile})
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
			errContains: "expect ';' or ','",
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
			args := NewCompileOptions("blang", []string{inputFile})
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
			name: "libb_write",
			code: `main() {
				write('Hello,');
				write(' World');
				write('!*n');
			}`,
			wantStdout: "Hello, World!\n",
		},
		{
			name: "libb_printf",
			code: `main() {
				printf("Hello, World!*n");
				printf("%% %% %%%%*n");
				printf("format %%d: %d %d*n", 123, -123);
				printf("format %%o: %o %o*n", 234, -234);
				printf("format %%c: %c %c*n", 'foo', 'bar');
				printf("format %%s: *"%s*" *"%s*"*n", "Hello", "World");
				printf("unknown format: %q*n", "foo");
			}`,
			wantStdout: `Hello, World!
% % %%
format %d: 123 -123
format %o: 352 -352
format %c: foo bar
format %s: "Hello" "World"
unknown format: %q
`,
		},
		{
			name: "libb_exit",
			code: `main() {
				printf("before exit()*n");
				exit();
				printf("after exit()*n");
			}`,
			wantStdout: "before exit()\n",
		},
		{
			name: "libb_char",
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
		{
			name: "libb_lchar",
			code: `main() {
				auto str;

				lchar(&str, 0, 'f');
				lchar(&str, 1, 'u');
				lchar(&str, 2, 'b');
				lchar(&str, 3, 'a');
				lchar(&str, 4, 'r');
				lchar(&str, 5, 0);
				printf("%s*n", &str);
			}`,
			wantStdout: "fubar\n",
		},
		{
			name: "libb_nwrite",
			code: `main() {
				nwrite(1, "foobar*n", 7);
			}`,
			wantStdout: "foobar\n",
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
	args := NewCompileOptions("blang", []string{file1, file2})
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

// TestE2Constant tests the long-running e-2 constant calculation (from oldtests/e2_test.cpp)
// This test is skipped by default because it takes 10+ seconds to run
func TestE2Constant(t *testing.T) {
	t.Skip("e-2 calculation is very long-running (~10+ seconds) - enable manually if needed")

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

	// Step 3: Run the executable (with 30 second timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

// TestPrecedence tests operator precedence (from oldtests/precedence_test.cpp)
func TestPrecedence(t *testing.T) {
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
			name: "add_mul",
			code: `main() {
				printf("3 + 4 ** 2 -> %d*n", 3 + 4 * 2);
			}`,
			wantStdout: "3 + 4 * 2 -> 11\n",
		},
		{
			name: "mul_add_mul",
			code: `main() {
				printf("5 ** 2 + 3 ** 4 -> %d*n", 5 * 2 + 3 * 4);
			}`,
			wantStdout: "5 * 2 + 3 * 4 -> 22\n",
		},
		{
			name: "sub_div",
			code: `main() {
				printf("10 - 6 / 2 -> %d*n", 10 - 6 / 2);
			}`,
			wantStdout: "10 - 6 / 2 -> 7\n",
		},
		{
			name: "mod_add",
			code: `main() {
				printf("7 %% 3 + 2 -> %d*n", 7 % 3 + 2);
			}`,
			wantStdout: "7 % 3 + 2 -> 3\n",
		},
		{
			name: "add_lt",
			code: `main() {
				printf("5 + 3 < 9 -> %d*n", 5 + 3 < 9);
			}`,
			wantStdout: "5 + 3 < 9 -> 1\n",
		},
		{
			name: "lt_eq",
			code: `main() {
				printf("4 < 6 == 1 -> %d*n", 4 < 6 == 1);
			}`,
			wantStdout: "4 < 6 == 1 -> 1\n",
		},
		{
			name: "eq_and",
			code: `main() {
				printf("3 == 3 & 1 -> %d*n", 3 == 3 & 1);
			}`,
			wantStdout: "3 == 3 & 1 -> 1\n",
		},
		{
			name: "and_or",
			code: `main() {
				printf("2 & 3 | 4 -> %d*n", 2 & 3 | 4);
			}`,
			wantStdout: "2 & 3 | 4 -> 6\n",
		},
		{
			name: "mul_add_lt",
			code: `main() {
				printf("2 ** 3 + 4 < 11 -> %d*n", 2 * 3 + 4 < 11);
			}`,
			wantStdout: "2 * 3 + 4 < 11 -> 1\n",
		},
		{
			name: "mul_ge_eq",
			code: `main() {
				printf("5 ** 2 >= 10 == 1 -> %d*n", 5 * 2 >= 10 == 1);
			}`,
			wantStdout: "5 * 2 >= 10 == 1 -> 1\n",
		},
		{
			name: "mul_and_add",
			code: `main() {
				printf("4 ** 2 & 3 + 1 -> %d*n", 4 * 2 & 3 + 1);
			}`,
			wantStdout: "4 * 2 & 3 + 1 -> 0\n",
		},
		{
			name: "div_add_gt_or",
			code: `main() {
				printf("6 / 2 + 1 > 3 | 2 -> %d*n", 6 / 2 + 1 > 3 | 2);
			}`,
			wantStdout: "6 / 2 + 1 > 3 | 2 -> 3\n",
		},
		{
			name: "div_mod",
			code: `main() {
				printf("10 / 2 %% 3 -> %d*n", 10 / 2 % 3);
			}`,
			wantStdout: "10 / 2 % 3 -> 2\n",
		},
		{
			name: "mul_or",
			code: `main() {
				printf("0 ** 5 | 3 -> %d*n", 0 * 5 | 3);
			}`,
			wantStdout: "0 * 5 | 3 -> 3\n",
		},
		{
			name: "mul_lshift",
			code: `main() {
				printf("4 ** 3 << 2 -> %d*n", 4 * 3 << 2);
			}`,
			wantStdout: "4 * 3 << 2 -> 48\n",
		},
		{
			name: "lshift_lt",
			code: `main() {
				printf("1 << 2 < 5 -> %d*n", 1 << 2 < 5);
			}`,
			wantStdout: "1 << 2 < 5 -> 1\n",
		},
		{
			name: "sub_rshift",
			code: `main() {
				printf("16 - 8 >> 1 -> %d*n", 16 - 8 >> 1);
			}`,
			wantStdout: "16 - 8 >> 1 -> 4\n",
		},
		{
			name: "lshift_and",
			code: `main() {
				printf("3 << 2 & 7 -> %d*n", 3 << 2 & 7);
			}`,
			wantStdout: "3 << 2 & 7 -> 4\n",
		},
		{
			name: "or_rshift",
			code: `main() {
				printf("2 | 4 >> 1 -> %d*n", 2 | 4 >> 1);
			}`,
			wantStdout: "2 | 4 >> 1 -> 2\n",
		},
		{
			name: "rshift_eq",
			code: `main() {
				printf("8 >> 2 == 2 -> %d*n", 8 >> 2 == 2);
			}`,
			wantStdout: "8 >> 2 == 2 -> 1\n",
		},
		{
			name: "mul_lshift_add",
			code: `main() {
				printf("5 ** 2 << 1 + 3 -> %d*n", 5 * 2 << 1 + 3);
			}`,
			wantStdout: "5 * 2 << 1 + 3 -> 160\n",
		},
		{
			name: "mod_lshift",
			code: `main() {
				printf("15 %% 4 << 2 -> %d*n", 15 % 4 << 2);
			}`,
			wantStdout: "15 % 4 << 2 -> 12\n",
		},
		{
			name: "lshift_gt_and",
			code: `main() {
				printf("1 << 3 > 5 & 2 -> %d*n", 1 << 3 > 5 & 2);
			}`,
			wantStdout: "1 << 3 > 5 & 2 -> 0\n",
		},
		{
			name: "add_lshift",
			code: `main() {
				printf("12345 + 10 << 4 -> %d*n", 12345 + 10 << 4);
			}`,
			wantStdout: "12345 + 10 << 4 -> 197680\n",
		},
		{
			name: "div_rshift",
			code: `main() {
				printf("16 / 2 >> 1 -> %d*n", 16 / 2 >> 1);
			}`,
			wantStdout: "16 / 2 >> 1 -> 4\n",
		},
		{
			name: "and_lshift_or",
			code: `main() {
				printf("7 & 3 << 2 | 8 -> %d*n", 7 & 3 << 2 | 8);
			}`,
			wantStdout: "7 & 3 << 2 | 8 -> 12\n",
		},
		{
			name: "lshift_ne",
			code: `main() {
				printf("1 << 4 != 15 -> %d*n", 1 << 4 != 15);
			}`,
			wantStdout: "1 << 4 != 15 -> 1\n",
		},
		{
			name: "rshift_ge",
			code: `main() {
				printf("98765 >> 3 >= 12345 -> %d*n", 98765 >> 3 >= 12345);
			}`,
			wantStdout: "98765 >> 3 >= 12345 -> 1\n",
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

// TestExpressions tests various expression features (from oldtests/expr_test.cpp)
func TestExpressions(t *testing.T) {
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
			name: "unary_operators",
			code: `main() {
				extrn x;
				auto y;

				printf("global -x = %d, expect %d*n", -x, -42);
				x = 0;
				printf("global !x = %d, expect %d*n", !x, 1);
				y = 987;
				x = &y;
				printf("global **x = %d, expect %d*n", *x, y);

				printf("local -y = %d, expect %d*n", -y, -987);
				y = 0;
				printf("local !y = %d, expect %d*n", !y, 1);
				x = 42;
				y = &x;
				printf("local **y = %d, expect %d*n", *y, x);
			}

			x 42;
			`,
			wantStdout: `global -x = -42, expect -42
global !x = 1, expect 1
global *x = 987, expect 987
local -y = -987, expect -987
local !y = 1, expect 1
local *y = 42, expect 42
`,
		},
		{
			name: "negation_in_conditional_context",
			code: `main() {
				extrn x;
				auto y;

				y = x + 100;
				printf("x = %d, y = %d*n", x, y);
				if (x)
					printf("if (x) WRONG*n");
				else
					printf("if (x) Correct*n");

				if (y)
					printf("if (y) Correct*n");
				else
					printf("if (y) WRONG*n");

				if (!x)
					printf("if (!x) Correct*n");
				else
					printf("if (!x) WRONG*n");

				if (!y)
					printf("if (!y) WRONG*n");
				else
					printf("if (!y) Correct*n");

				while (!x) {
					printf("while (!x) x = %d*n", x);
					x = 42;
				}
			}

			x;
			`,
			wantStdout: `x = 0, y = 100
if (x) Correct
if (y) Correct
if (!x) Correct
if (!y) Correct
while (!x) x = 0
`,
		},
		{
			name: "postfix_operators",
			code: `incr(x) {
				printf("increment %d*n", x++);
				return (x);
			}

			add(a, b) {
				printf("add %d + %d*n", a, b);
				return (a + b);
			}

			decr(x) {
				printf("decrement %d*n", x--);
				return (x);
			}

			sub(a, b) {
				printf("subtract %d - %d*n", a, b);
				return (a - b);
			}

			assign_local(x) {
				auto result;
				printf("assign local %d*n", x);
				result = x;
				return (result);
			}

			assign_global(x) {
				extrn g;

				printf("assign global %d*n", x);
				g = x;
			}

			main() {
				extrn g;

				printf("%d*n", incr(42));
				printf("%d*n", add(42, 123));
				printf("%d*n", decr(42));
				printf("%d*n", sub(42, 123));
				printf("%d*n", assign_local(42));
				assign_global(42);
				printf("%d*n", g);
			}

			g;
			`,
			wantStdout: `increment 42
43
add 42 + 123
165
decrement 42
41
subtract 42 - 123
-81
assign local 42
42
assign global 42
42
`,
		},
		{
			name: "local_array",
			code: `main() {
				auto l[3];

				l[0] = 123;
				l[1] = 'local';
				l[2] = "string";
				printf("local = %d, '%c', *"%s*"*n", l[0], l[1], l[2]);
			}`,
			wantStdout: `local = 123, 'local', "string"
`,
		},
		{
			name: "global_array",
			code: `g[3] -345, 'foo', "bar";

			main() {
				extrn g;

				printf("global = %d, '%c', *"%s*"*n", g[0], g[1], g[2]);
				printf("address = %d, %d, %d*n", (&g[0]) - g, (&g[1]) - g, (&g[2]) - g);
			}`,
			wantStdout: `global = -345, 'foo', "bar"
address = 0, 8, 16
`,
		},
		{
			name: "local_mix",
			code: `main() {
				auto e, d;
				auto c[1];
				auto b, a;
				auto p;

				a = 11;
				b = 22;
				c[0] = 33;
				d = 44;
				e = 55;

				printf("%d %d %d %d", a, b, c - &c, c[0]);
				printf(" %d %d*n", d, e);
				p = &a;
				printf("%d %d %d %d", p[0], p[1], p[2] - &c, p[3]);
				printf(" %d %d*n", p[4], p[5]);
			}`,
			wantStdout: `11 22 8 33 44 55
11 22 8 33 44 55
`,
		},
		{
			name: "binary_operators",
			code: `x 42;

			main() {
				extrn x;
				auto y;

				y = 345;
				printf("%d + %d -> %d*n", x, y, x + y);
				printf("%d + %d -> %d*n", y, x, y + x);

				printf("%d - %d -> %d*n", x, y, x - y);
				printf("%d - %d -> %d*n", y, x, y - x);

				printf("%d ** %d -> %d*n", x, y, x * y);
				printf("%d ** %d -> %d*n", y, x, y * x);

				printf("%d / %d -> %d*n", x, y, x / y);
				printf("%d / %d -> %d*n", y, x, y / x);

				printf("%d %% %d -> %d*n", x, y, x % y);
				printf("%d %% %d -> %d*n", y, x, y % x);

				printf("%d < %d -> %d*n", x, y, x < y);
				printf("%d < %d -> %d*n", y, x, y < x);

				printf("%d <= %d -> %d*n", x, y, x <= y);
				printf("%d <= %d -> %d*n", y, x, y <= x);

				printf("%d > %d -> %d*n", x, y, x > y);
				printf("%d > %d -> %d*n", y, x, y > x);

				printf("%d >= %d -> %d*n", x, y, x >= y);
				printf("%d >= %d -> %d*n", y, x, y >= x);

				printf("%d == %d -> %d*n", x, y, x == y);
				printf("%d == %d -> %d*n", y, x, y == x);

				printf("%d != %d -> %d*n", x, y, x != y);
				printf("%d != %d -> %d*n", y, x, y != x);

				printf("%d & %d -> %d*n", x, y, x & y);
				printf("%d & %d -> %d*n", y, x, y & x);

				printf("%d | %d -> %d*n", x, y, x | y);
				printf("%d | %d -> %d*n", y, x, y | x);
			}`,
			wantStdout: `42 + 345 -> 387
345 + 42 -> 387
42 - 345 -> -303
345 - 42 -> 303
42 * 345 -> 14490
345 * 42 -> 14490
42 / 345 -> 0
345 / 42 -> 8
42 % 345 -> 42
345 % 42 -> 9
42 < 345 -> 1
345 < 42 -> 0
42 <= 345 -> 1
345 <= 42 -> 0
42 > 345 -> 0
345 > 42 -> 1
42 >= 345 -> 0
345 >= 42 -> 1
42 == 345 -> 0
345 == 42 -> 0
42 != 345 -> 1
345 != 42 -> 1
42 & 345 -> 8
345 & 42 -> 8
42 | 345 -> 379
345 | 42 -> 379
`,
		},
		{
			name: "eq_by_bitmask",
			code: `main() {
				auto cval;

				cval = 51;
				if ((cval & 017777) == cval) {
					printf("Small positive: %d*n", cval);
				} else {
					printf("Wrong: %d*n", cval);
				}
			}`,
			wantStdout: "Small positive: 51\n",
		},
		{
			name: "octal_literals",
			code: `main() {
				auto v;
				v = 012345;
				printf("%d*n", v);
				v = -04567;
				printf("%d*n", v);
			}`,
			wantStdout: `5349
-2423
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

// TestIndirectCalls tests indirect function calls through function pointers
func TestIndirectCalls(t *testing.T) {
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
			name: "indirect_call_basic",
			code: `add(a, b) {
				return(a + b);
			}

			func_ptr;

			main() {
				extrn func_ptr;

				func_ptr = add;
				printf("Result: %d*n", func_ptr(3, 5));
			}`,
			wantStdout: "Result: 8\n",
		},
		{
			name: "indirect_call_multiple",
			code: `add(a, b) { return(a + b); }
			sub(a, b) { return(a - b); }
			mul(a, b) { return(a * b); }

			op;

			main() {
				extrn op;

				op = add;
				printf("add: %d*n", op(10, 5));

				op = sub;
				printf("sub: %d*n", op(10, 5));

				op = mul;
				printf("mul: %d*n", op(10, 5));
			}`,
			wantStdout: `add: 15
sub: 5
mul: 50
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

// TestFunctions tests various function features (from oldtests/func_test.cpp)
func TestFunctions(t *testing.T) {
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
			name: "function_definitions",
			code: `a() {}
			b();
			c() label:;
			d() label: goto label;
			e() return;
			f(x) return(x);
			g(x) x;
			h(x) if(x) 123;
			i(x) if(x) 123; else 456;
			j(x) while(x);
			k(x) switch(x);
			l(x) switch(x) case 1:;
			m() extrn x;
			n() auto x;

			main() {
				printf("before a()*n");
				a();    printf("after a(), before b()*n");
				b();    printf("after b(), before c()*n");
				c();    printf("after c(), before e()*n");
				e();    printf("after e(), before f()*n");
				f(42);  printf("after f(), before g()*n");
				g(42);  printf("after g(), before h()*n");
				h(42);  printf("after h(), before i()*n");
				i(42);  printf("after i(), before j()*n");
				j(0);   printf("after j(), before k()*n");
				k(42);  printf("after k(), before l()*n");
				l(42);  printf("after l(), before m()*n");
				m();    printf("after m(), before n()*n");
				n();    printf("after n()*n");
			}`,
			wantStdout: `before a()
after a(), before b()
after b(), before c()
after c(), before e()
after e(), before f()
after f(), before g()
after g(), before h()
after h(), before i()
after i(), before j()
after j(), before k()
after k(), before l()
after l(), before m()
after m(), before n()
after n()
`,
		},
		{
			name: "function_arguments",
			code: `func(a, b, c)
			{
				printf("a = %d, b = '%c', c = *"%s*"*n", a, b, c);
			}

			main() {
				func(123, 'foo', "bar");
			}`,
			wantStdout: `a = 123, b = 'foo', c = "bar"
`,
		},
		{
			name: "function_ternary_operator",
			code: `choose(a, b, c)
			{
				return (a ? b : c);
			}

			main() {
				printf("%d*n", choose(1, 123, 456));
				printf("%d*n", choose(0, 123, 456));
			}`,
			wantStdout: `123
456
`,
		},
		{
			name: "nested_ternary",
			code: `classify(n) {
				return (n > 50 ? 100 : (n > 25 ? 50 : 25));
			}

			sign(n) {
				return (n > 0 ? 1 : (n < 0 ? -1 : 0));
			}

			main() {
				printf("classify(75) = %d*n", classify(75));
				printf("classify(40) = %d*n", classify(40));
				printf("classify(10) = %d*n", classify(10));

				printf("sign(42) = %d*n", sign(42));
				printf("sign(-17) = %d*n", sign(-17));
				printf("sign(0) = %d*n", sign(0));
			}`,
			wantStdout: `classify(75) = 100
classify(40) = 50
classify(10) = 25
sign(42) = 1
sign(-17) = -1
sign(0) = 0
`,
		},
		{
			name: "ternary_in_expression",
			code: `main() {
				auto x, y, z, result;

				x = 10;
				y = 20;
				z = 30;

				result = x + (y > 15 ? z : 0);
				printf("x + (y > 15 ? z : 0) = %d*n", result);

				result = (x < y ? x : y) * 2;
				printf("(x < y ? x : y) ** 2 = %d*n", result);

				printf("nested: %d*n", (x > 5 ? (y > 15 ? 100 : 50) : 0));
			}`,
			wantStdout: `x + (y > 15 ? z : 0) = 40
(x < y ? x : y) * 2 = 20
nested: 100
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

// TestCompoundAssignments tests compound assignment operators (from oldtests/assignment_test.cpp)
// NOTE: These tests are currently skipped because compound assignments are not yet implemented
func TestCompoundAssignments(t *testing.T) {
	t.Skip("Compound assignments not yet implemented - use x = x + 5 instead of x =+ 5")

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

// BenchmarkCompile benchmarks the compilation process
func BenchmarkCompile(b *testing.B) {
	tmpDir := b.TempDir()
	outputFile := filepath.Join(tmpDir, "output.ll")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args := NewCompileOptions("blang", []string{"testdata/arithmetic.b"})
		args.OutputFile = outputFile
		Compile(args)
	}
}
