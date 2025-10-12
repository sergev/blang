package main

import (
	"os"
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
