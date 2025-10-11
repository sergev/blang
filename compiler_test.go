package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCompile tests the full compilation pipeline
func TestCompile(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		expectFile string
	}{
		{"hello", "testdata/hello.b", "testdata/expected/hello.s"},
		{"arithmetic", "testdata/arithmetic.b", "testdata/expected/arithmetic.s"},
		{"globals", "testdata/globals.b", "testdata/expected/globals.s"},
		{"conditionals", "testdata/conditionals.b", "testdata/expected/conditionals.s"},
		{"loops", "testdata/loops.b", "testdata/expected/loops.s"},
		{"strings", "testdata/strings.b", "testdata/expected/strings.s"},
		{"operators", "testdata/operators.b", "testdata/expected/operators.s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary output file
			outputFile := filepath.Join(t.TempDir(), "output.s")

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

			// Read expected output
			want, err := os.ReadFile(tt.expectFile)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			// Compare outputs
			if string(got) != string(want) {
				t.Errorf("Output mismatch for %s\nGot:\n%s\nWant:\n%s",
					tt.name, string(got), string(want))
			}
		})
	}
}

// TestCompileErrors tests that invalid B programs are rejected
// Note: These tests are skipped because the compiler calls os.Exit()
// on errors rather than returning an error. To properly test error handling,
// we would need to refactor the error handling or run tests in subprocesses.
func TestCompileErrors(t *testing.T) {
	t.Skip("Error handling tests skipped - compiler calls os.Exit() on errors")

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "undefined_variable",
			content: "main() { x = 10; }",
			wantErr: true,
		},
		{
			name:    "unclosed_comment",
			content: "/* unclosed comment\nmain() { }",
			wantErr: true,
		},
		{
			name:    "missing_semicolon",
			content: "main() { auto x x = 10; }",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary input file
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test.b")
			outputFile := filepath.Join(tmpDir, "output.s")

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
	err = os.WriteFile(file2, []byte("main() { extrn add; return(add(1, 2)); }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Compile both files
	outputFile := filepath.Join(tmpDir, "output.s")
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
}

// BenchmarkCompile benchmarks the compilation process
func BenchmarkCompile(b *testing.B) {
	tmpDir := b.TempDir()
	outputFile := filepath.Join(tmpDir, "output.s")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args := NewCompilerArgs("blang", []string{"testdata/arithmetic.b"})
		args.OutputFile = outputFile
		Compile(args)
	}
}
