package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIBasicOptions tests basic CLI options like help, version, and output
func TestCLIBasicOptions(t *testing.T) {
	ensureBlangOrSkip(t)
	tests := []struct {
		name        string
		args        []string
		wantExit    int
		wantStdout  string
		wantStderr  string
		skipOnError bool
	}{
		{
			name:       "help_option",
			args:       []string{"--help"},
			wantExit:   0,
			wantStdout: "Usage: blang [options] file...",
		},
		{
			name:       "version_option",
			args:       []string{"--version"},
			wantExit:   0,
			wantStdout: "blang version 0.1",
		},
		{
			name:       "no_input_files",
			args:       []string{},
			wantExit:   1,
			wantStdout: "Usage: blang [options] file...",
		},
		{
			name:       "invalid_file_extension",
			args:       []string{"test.txt"},
			wantExit:   1,
			wantStderr: "unsupported input file extension",
		},
		{
			name:       "nonexistent_file",
			args:       []string{"nonexistent.b"},
			wantExit:   1,
			wantStderr: "cannot access file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
			}

			outputStr := string(output)
			if tt.wantStdout != "" && !strings.Contains(outputStr, tt.wantStdout) {
				t.Errorf("Output doesn't contain expected stdout: %q", tt.wantStdout)
			}
			if tt.wantStderr != "" && !strings.Contains(outputStr, tt.wantStderr) {
				t.Errorf("Output doesn't contain expected stderr: %q", tt.wantStderr)
			}
		})
	}
}

// TestCLIOutputFormats tests different output format options
func TestCLIOutputFormats(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Hello*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		wantExit   int
		wantOutput string
		checkFile  bool
		fileExt    string
	}{
		{
			name:       "default_executable",
			args:       []string{"-L", "runtime", "-o", filepath.Join(tmpDir, "test_default"), testFile},
			wantExit:   0,
			wantOutput: "",
			checkFile:  true,
			fileExt:    "",
		},
		{
			name:       "llvm_ir_output",
			args:       []string{"--emit-llvm", "-o", filepath.Join(tmpDir, "test.ll"), testFile},
			wantExit:   0,
			wantOutput: "",
			checkFile:  true,
			fileExt:    ".ll",
		},
		{
			name:       "object_file_output",
			args:       []string{"-c", "-o", filepath.Join(tmpDir, "test.o"), testFile},
			wantExit:   0,
			wantOutput: "",
			checkFile:  true,
			fileExt:    ".o",
		},
		{
			name:       "assembly_output",
			args:       []string{"-S", "-o", filepath.Join(tmpDir, "test.s"), testFile},
			wantExit:   0,
			wantOutput: "",
			checkFile:  true,
			fileExt:    ".s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}

			if tt.checkFile {
				outputFile := tt.args[len(tt.args)-2] // -o output_file
				if _, err := os.Stat(outputFile); os.IsNotExist(err) {
					t.Errorf("Output file %s was not created", outputFile)
				} else {
					// Check file extension if specified
					if tt.fileExt != "" && !strings.HasSuffix(outputFile, tt.fileExt) {
						t.Errorf("Output file %s does not have expected extension %s", outputFile, tt.fileExt)
					}
				}
			}
		})
	}
}

// TestCLIOptimizationFlags tests optimization level flags
func TestCLIOptimizationFlags(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    auto x;
    x = 42;
    return(x);
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "optimization_O0",
			args:     []string{"-L", "runtime", "-O0", "-o", filepath.Join(tmpDir, "test_O0"), testFile},
			wantExit: 0,
		},
		{
			name:     "optimization_O1",
			args:     []string{"-L", "runtime", "-O1", "-o", filepath.Join(tmpDir, "test_O1"), testFile},
			wantExit: 0,
		},
		{
			name:     "optimization_O2",
			args:     []string{"-L", "runtime", "-O2", "-o", filepath.Join(tmpDir, "test_O2"), testFile},
			wantExit: 0,
		},
		{
			name:     "optimization_O3",
			args:     []string{"-L", "runtime", "-O3", "-o", filepath.Join(tmpDir, "test_O3"), testFile},
			wantExit: 0,
		},
		{
			name:     "invalid_optimization",
			args:     []string{"-O4", testFile},
			wantExit: 1, // pflag correctly detects invalid optimization level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}
		})
	}
}

// TestCLIDebugAndVerbose tests debug and verbose flags
func TestCLIDebugAndVerbose(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Test*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		wantExit   int
		wantOutput string
	}{
		{
			name:       "verbose_output",
			args:       []string{"-L", "runtime", "-v", "-o", filepath.Join(tmpDir, "test_verbose"), testFile},
			wantExit:   0,
			wantOutput: "blang: compiling",
		},
		{
			name:       "debug_info",
			args:       []string{"-L", "runtime", "-g", "-o", filepath.Join(tmpDir, "test_debug"), testFile},
			wantExit:   0,
			wantOutput: "",
		},
		{
			name:       "verbose_and_debug",
			args:       []string{"-L", "runtime", "-v", "-g", "-o", filepath.Join(tmpDir, "test_verbose_debug"), testFile},
			wantExit:   0,
			wantOutput: "blang: compiling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}

			if tt.wantOutput != "" {
				outputStr := string(output)
				if !strings.Contains(outputStr, tt.wantOutput) {
					t.Errorf("Output doesn't contain expected text: %q", tt.wantOutput)
					t.Logf("Full output: %s", outputStr)
				}
			}
		})
	}
}

// TestCLIWarningFlags tests warning flags
func TestCLIWarningFlags(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Test*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}
		})
	}
}

// TestCLIPathFlags tests include and library path flags
func TestCLIPathFlags(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Test*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "library_path",
			args:     []string{"-L", "/tmp", "-o", filepath.Join(tmpDir, "test_libpath"), testFile},
			wantExit: 0,
		},
		{
			name:     "library_link",
			args:     []string{"-L", "runtime", "-l", "c", "-o", filepath.Join(tmpDir, "test_lib"), testFile},
			wantExit: 0,
		},
		{
			name:     "multiple_library_dirs",
			args:     []string{"-L", "runtime", "-L", "/usr/lib", "-L", "/usr/local/lib", "-o", filepath.Join(tmpDir, "test_multi_libpath"), testFile},
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}
		})
	}
}

// TestCLIStandardFlag tests the language standard flag
func TestCLIStandardFlag(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Test*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "default_standard",
			args:     []string{"-L", "runtime", "-o", filepath.Join(tmpDir, "test_default_std"), testFile},
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}
		})
	}
}

// TestCLISaveTemps tests the save-temps flag
func TestCLISaveTemps(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Test*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		wantExit       int
		expectTempFile bool
	}{
		{
			name:           "without_save_temps",
			args:           []string{"-L", "runtime", "-o", filepath.Join(tmpDir, "test_no_save"), testFile},
			wantExit:       0,
			expectTempFile: false,
		},
		{
			name:           "with_save_temps",
			args:           []string{"-L", "runtime", "--save-temps", "-o", filepath.Join(tmpDir, "test_save"), testFile},
			wantExit:       0,
			expectTempFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}

			// Check for temporary files
			outputFile := tt.args[len(tt.args)-2] // -o output_file
			tempFile := outputFile + ".tmp.ll"
			if _, err := os.Stat(tempFile); tt.expectTempFile {
				if os.IsNotExist(err) {
					t.Errorf("Expected temporary file %s to exist when --save-temps is used", tempFile)
				}
			} else {
				if !os.IsNotExist(err) {
					t.Errorf("Expected temporary file %s to be cleaned up when --save-temps is not used", tempFile)
				}
			}
		})
	}
}

// TestCLIExecutableGeneration tests that generated executables actually work
func TestCLIExecutableGeneration(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Hello*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "test_executable")

	// Compile the executable
	cmd := exec.Command("./blang", "-L", "runtime", "-o", outputFile, testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compilation failed: %v\nOutput: %s", err, string(output))
	}

	// Make sure the file exists and is executable
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Executable was not created: %s", outputFile)
	}

	// Test running the executable
	runCmd := exec.Command(outputFile)
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Executable failed to run: %v\nOutput: %s", err, string(runOutput))
	}

	expectedOutput := "Hello"
	if !strings.Contains(string(runOutput), expectedOutput) {
		t.Errorf("Executable output doesn't contain expected text: %q\nActual output: %s", expectedOutput, string(runOutput))
	}
}

// TestCLICombinedFlags tests combinations of multiple flags
func TestCLICombinedFlags(t *testing.T) {
	ensureBlangOrSkip(t)
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.b")

	// Create a simple test file
	testCode := `main() {
    write('Test*n');
}`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "verbose_optimized_debug",
			args:     []string{"-L", "runtime", "-v", "-O2", "-g", "-o", filepath.Join(tmpDir, "test_combined1"), testFile},
			wantExit: 0,
		},
		{
			name:     "verbose_optimized_debug_O1",
			args:     []string{"-L", "runtime", "-v", "-O1", "-g", "-o", filepath.Join(tmpDir, "test_combined2"), testFile},
			wantExit: 0,
		},
		{
			name:     "all_flags",
			args:     []string{"-L", "runtime", "-v", "-O3", "-g", "--save-temps", "-o", filepath.Join(tmpDir, "test_combined3"), testFile},
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./blang", tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Command failed with non-exit error: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
				t.Logf("Command output: %s", string(output))
			}
		})
	}
}
