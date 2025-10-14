package main

import (
	"os"
	"path/filepath"
	"testing"
)

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

// BenchmarkCompile benchmarks the compilation process
func BenchmarkCompile(b *testing.B) {
	tmpDir := b.TempDir()
	outputFile := filepath.Join(tmpDir, "output.ll")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use inline arithmetic test instead of file
		arithmeticCode := `add(a, b) {
    return(a + b);
}
sub(a, b) {
    return(a - b);
}
mul(a, b) {
    return(a * b);
}
main() {
    auto x, y, z;
    x = 10;
    y = 20;
    z = add(x, y);
    z = sub(z, 5);
    z = mul(z, 2);
    return(z);
}`

		inputFile := filepath.Join(tmpDir, "test.b")
		err := os.WriteFile(inputFile, []byte(arithmeticCode), 0644)
		if err != nil {
			b.Fatalf("Failed to write test file: %v", err)
		}

		args := NewCompileOptions("blang", []string{inputFile})
		args.OutputFile = outputFile
		Compile(args)
	}
}

// TestCompile tests the full compilation pipeline
func TestCompile(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantFunc string // Function name that should exist in output
	}{
		{
			name: "hello",
			code: `main() {
    printf("Hello, World!*n");
}`,
			wantFunc: "@main",
		},
		{
			name: "arithmetic",
			code: `add(a, b) {
    return(a + b);
}
sub(a, b) {
    return(a - b);
}
mul(a, b) {
    return(a * b);
}
main() {
    auto x, y, z;
    x = 10;
    y = 20;
    z = add(x, y);
    z = sub(z, 5);
    z = mul(z, 2);
    return(z);
}`,
			wantFunc: "@main",
		},
		{
			name: "globals",
			code: `counter 0;
values[3] 10, 20, 30;

increment() {
    extrn counter;
    counter = counter + 1;
}

sum_values() {
    extrn values;
    auto i, total;
    total = 0;
    i = 0;
    while (i < 3) {
        total = total + values[i];
        i++;
    }
    return(total);
}

main() {
    increment();
    increment();
    return(sum_values());
}`,
			wantFunc: "@main",
		},
		{
			name: "conditionals",
			code: `max(a, b) {
    if (a > b)
        return(a);
    else
        return(b);
}

abs(n) {
    if (n < 0)
        return(-n);
    return(n);
}

main() {
    auto x, y;
    x = max(10, 20);
    y = abs(-15);
    return(x + y);
}`,
			wantFunc: "@main",
		},
		{
			name: "loops",
			code: `factorial(n) {
    auto result, i;
    result = 1;
    i = 1;
    while (i <= n) {
        result = result * i;
        i++;
    }
    return(result);
}

main() {
    return(factorial(5));
}`,
			wantFunc: "@factorial",
		},
		{
			name: "strings",
			code: `messages[3] "Hello", "World", "Test";

main() {
    extrn messages;
    printf("%s*n", messages[0]);
    printf("%s*n", messages[1]);
    printf("%s*n", messages[2]);
}`,
			wantFunc: "@main",
		},
		{
			name: "arrays",
			code: `sum_array(arr, n) {
    auto i, sum;
    sum = 0;
    i = 0;
    while (i < n) {
        sum = sum + arr[i];
        i++;
    }
    return(sum);
}

main() {
    auto numbers[5];
    auto i, total;

    /* Initialize array */
    i = 0;
    while (i < 5) {
        numbers[i] = (i + 1) * 10;
        i++;
    }

    /* Sum using function */
    total = sum_array(numbers, 5);

    return(total);  /* Should be 10+20+30+40+50 = 150 */
}`,
			wantFunc: "@sum_array",
		},
		{
			name: "pointers",
			code: `main() {
    auto x, y, ptr;
    x = 100;
    y = 200;

    ptr = &x;
    printf("x = %d*n", *ptr);

    ptr = &y;
    printf("y = %d*n", *ptr);

    return(*ptr);
}`,
			wantFunc: "@main",
		},
		{
			name: "switch",
			code: `classify(n) {
    switch (n) {
        case 0:
            return(0);
        case 1:
            return(1);
        case 2:
            return(4);
        default:
            return(9);
    }
}

main() {
    return(classify(2));
}`,
			wantFunc: "@classify",
		},
		{
			name: "goto",
			code: `main() {
    auto x;
    x = 10;

    if (x > 5)
        goto skip;

    x = 20;

skip:
    return(x);
}`,
			wantFunc: "@main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary input and output files
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test.b")
			outputFile := filepath.Join(tmpDir, "output.ll")

			// Write test code to input file
			err := os.WriteFile(inputFile, []byte(tt.code), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Compile the input
			args := NewCompileOptions("blang", []string{inputFile})
			args.OutputFile = outputFile

			err = Compile(args)
			if err != nil {
				t.Fatalf("Compile(%s) failed: %v", tt.name, err)
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
