package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfig holds configuration for a test case
type TestConfig struct {
	Name           string
	Args           []string
	WantExit       int
	WantStdout     string
	WantStderr     string
	WantOutput     string
	CheckFile      bool
	FileExt        string
	ExpectTempFile bool
}

// CompileTestConfig holds configuration for compilation tests
type CompileTestConfig struct {
	Name       string
	Code       string
	WantFunc   string // Function name that should exist in output
	WantStdout string
	WantExit   int
}

// LexerTestConfig holds configuration for lexer tests
type LexerTestConfig struct {
	Name  string
	Input string
	Want  interface{} // Can be string, int64, Token, etc.
}

// runBlangTest executes a blang command and validates the results
func runBlangTest(t *testing.T, config TestConfig) {
	t.Helper()

	cmd := exec.Command("./blang", config.Args...)
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Command failed with non-exit error: %v", err)
		}
	}

	if exitCode != config.WantExit {
		t.Errorf("Exit code = %d, want %d", exitCode, config.WantExit)
		t.Logf("Command output: %s", string(output))
	}

	outputStr := string(output)
	if config.WantStdout != "" && !strings.Contains(outputStr, config.WantStdout) {
		t.Errorf("Output doesn't contain expected stdout: %q", config.WantStdout)
	}
	if config.WantStderr != "" && !strings.Contains(outputStr, config.WantStderr) {
		t.Errorf("Output doesn't contain expected stderr: %q", config.WantStderr)
	}
	if config.WantOutput != "" && !strings.Contains(outputStr, config.WantOutput) {
		t.Errorf("Output doesn't contain expected text: %q", config.WantOutput)
		t.Logf("Full output: %s", outputStr)
	}

	if config.CheckFile {
		outputFile := config.Args[len(config.Args)-2] // -o output_file
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file %s was not created", outputFile)
		} else {
			// Check file extension if specified
			if config.FileExt != "" && !strings.HasSuffix(outputFile, config.FileExt) {
				t.Errorf("Output file %s does not have expected extension %s", outputFile, config.FileExt)
			}
		}
	}

	if config.ExpectTempFile {
		outputFile := config.Args[len(config.Args)-2] // -o output_file
		tempFile := outputFile + ".tmp.ll"
		if _, err := os.Stat(tempFile); os.IsNotExist(err) {
			t.Errorf("Expected temporary file %s to exist when -save-temps is used", tempFile)
		}
	} else if !config.ExpectTempFile && len(config.Args) > 2 {
		outputFile := config.Args[len(config.Args)-2] // -o output_file
		tempFile := outputFile + ".tmp.ll"
		if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
			t.Errorf("Expected temporary file %s to be cleaned up when -save-temps is not used", tempFile)
		}
	}
}

// runCompileTest compiles code to LLVM IR and validates the result
func runCompileTest(t *testing.T, config CompileTestConfig) {
	t.Helper()

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", config.Code)
	outputFile := filepath.Join(tmpDir, "output.ll")

	// Compile to LLVM IR for testing
	args := NewCompileOptions("test", []string{testFile})
	args.OutputFile = outputFile
	args.OutputType = OutputIR // Force LLVM IR output for testing
	err := Compile(args)

	if config.WantExit == 0 && err != nil {
		t.Errorf("Compilation failed: %v", err)
		return
	}

	if config.WantExit != 0 && err == nil {
		t.Errorf("Expected compilation to fail, but it succeeded")
		return
	}

	// Check for expected function in output
	if config.WantFunc != "" {
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Errorf("Failed to read output file: %v", err)
			return
		}

		if !strings.Contains(string(content), config.WantFunc) {
			t.Errorf("Output doesn't contain expected function %q", config.WantFunc)
			t.Logf("Output content: %s", string(content))
		}
	}

	// If we expect stdout, compile and run the program
	if config.WantStdout != "" {
		runCompileAndExecuteTest(t, config)
	}
}

// runCompileAndExecuteTest compiles code and executes it to check stdout
func runCompileAndExecuteTest(t *testing.T, config CompileTestConfig) {
	t.Helper()

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", config.Code)
	outputFile := filepath.Join(tmpDir, "test_executable")

	// Compile to executable
	cmd := exec.Command("./blang", "-o", outputFile, testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compilation failed: %v\nOutput: %s", err, string(output))
	}

	// Execute the program
	runCmd := exec.Command(outputFile)
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, string(runOutput))
	}

	// Check output
	if !strings.Contains(string(runOutput), config.WantStdout) {
		t.Errorf("Program output doesn't contain expected text: %q\nActual output: %s", config.WantStdout, string(runOutput))
	}
}

// runLexerTest tests lexer functionality
func runLexerTest(t *testing.T, config LexerTestConfig, testFunc func(*Lexer) (interface{}, error)) {
	t.Helper()

	args := NewCompileOptions("test", nil)
	lexer := NewLexer(args, strings.NewReader(config.Input))

	got, err := testFunc(lexer)
	if err != nil {
		t.Fatalf("Lexer test failed: %v", err)
	}

	if got != config.Want {
		t.Errorf("Lexer test = %v, want %v", got, config.Want)
	}
}

// createTestFile creates a simple B test file in the given directory
func createTestFile(t *testing.T, tmpDir, filename, code string) string {
	t.Helper()

	testFile := filepath.Join(tmpDir, filename)
	err := os.WriteFile(testFile, []byte(code), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return testFile
}

// requireLibbO skips the test if libb.o is not available
func requireLibbO(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("libb.o"); err != nil {
		t.Skip("libb.o not found, run 'make' first")
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
