package main

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkCompile benchmarks the compilation process
func BenchmarkCompile(b *testing.B) {
	tmpDir := b.TempDir()
	outputFile := filepath.Join(tmpDir, "output.ll")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args := NewCompileOptions("blang", []string{"examples/e-2.b"})
		args.OutputFile = outputFile
		Compile(args)
	}
}

// TestCompile tests the full compilation pipeline
func TestCompile(t *testing.T) {
	tests := []CompileTestConfig{
		{
			Name:     "hello",
			Code:     `main() { printf("Hello, World!*n"); }`,
			WantFunc: "@main",
		},
		{
			Name: "arithmetic",
			Code: `add(a, b) {
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
			WantFunc: "@main",
		},
		{
			Name: "globals",
			Code: `counter 0;
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
			WantFunc: "@main",
		},
		{
			Name: "conditionals",
			Code: `max(a, b) {
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
			WantFunc: "@main",
		},
		{
			Name: "loops",
			Code: `factorial(n) {
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
			WantFunc: "@factorial",
		},
		{
			Name: "strings",
			Code: `messages[3] "Hello", "World", "Test";

main() {
    extrn messages;
    printf("%s*n", messages[0]);
    printf("%s*n", messages[1]);
    printf("%s*n", messages[2]);
}`,
			WantFunc: "@main",
		},
		{
			Name: "arrays",
			Code: `sum_array(arr, n) {
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
			WantFunc: "@sum_array",
		},
		{
			Name: "pointers",
			Code: `main() {
    auto x, y, ptr;
    x = 100;
    y = 200;

    ptr = &x;
    printf("x = %d*n", *ptr);

    ptr = &y;
    printf("y = %d*n", *ptr);

    return(*ptr);
}`,
			WantFunc: "@main",
		},
		{
			Name: "switch",
			Code: `classify(n) {
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
			WantFunc: "@classify",
		},
		{
			Name: "goto",
			Code: `main() {
    auto x;
    x = 10;

    if (x > 5)
        goto skip;

    x = 20;

skip:
    return(x);
}`,
			WantFunc: "@main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runCompileTest(t, tt)
		})
	}
}

// TestCompileErrors tests that invalid B programs are rejected
func TestCompileErrors(t *testing.T) {
	tests := []CompileTestConfig{
		{
			Name:     "undefined_variable",
			Code:     "main() { x = 10; }",
			WantExit: 1,
		},
		{
			Name:     "unclosed_comment",
			Code:     "/* unclosed comment\nmain() { }",
			WantExit: 1,
		},
		{
			Name:     "missing_semicolon",
			Code:     "main() { auto x x = 10; }",
			WantExit: 1,
		},
		{
			Name:     "unclosed_char_literal",
			Code:     "main() { auto c; c = 'abcdefghij; }",
			WantExit: 1,
		},
		{
			Name:     "unterminated_string",
			Code:     `main() { write("hello); }`,
			WantExit: 1,
		},
		{
			Name:     "undefined_escape",
			Code:     `main() { write("*x"); }`,
			WantExit: 1,
		},
		{
			Name:     "case_outside_switch",
			Code:     "main() { case 1: return(0); }",
			WantExit: 1,
		},
		// Note: Duplicate identifier detection is not yet implemented in LLVM backend
		// {
		// 	Name:     "duplicate_identifier",
		// 	Code:     "main() { auto x, x; }",
		// 	WantExit: 1,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runCompileTest(t, tt)
		})
	}
}

// TestCompileMultipleFiles tests compiling multiple B files
func TestCompileMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first file
	file1 := createTestFile(t, tmpDir, "file1.b", "add(a, b) { return(a + b); }")
	// Create second file
	file2 := createTestFile(t, tmpDir, "file2.b", "main() { return(add(1, 2)); }")

	// Compile both files
	outputFile := filepath.Join(tmpDir, "output.ll")
	args := NewCompileOptions("blang", []string{file1, file2})
	args.OutputFile = outputFile
	args.OutputType = OutputIR

	err := Compile(args)
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
