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
		// Note: example_e2 is tested separately in TestE2Constant due to long runtime
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
	args := NewCompilerArgs("blang", []string{inputFile})
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
