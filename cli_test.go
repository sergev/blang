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
	tests := []TestConfig{
		{
			Name:       "help_option",
			Args:       []string{"--help"},
			WantExit:   0,
			WantStdout: "Usage: blang [options] file...",
		},
		{
			Name:       "version_option",
			Args:       []string{"--version"},
			WantExit:   0,
			WantStdout: "blang version 0.1",
		},
		{
			Name:       "no_input_files",
			Args:       []string{},
			WantExit:   1,
			WantStdout: "Usage: blang [options] file...",
		},
		{
			Name:       "invalid_file_extension",
			Args:       []string{"test.txt"},
			WantExit:   1,
			WantStderr: "does not have .b extension",
		},
		{
			Name:       "nonexistent_file",
			Args:       []string{"nonexistent.b"},
			WantExit:   1,
			WantStderr: "cannot access file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}

// TestCLIOutputFormats tests different output format options
func TestCLIOutputFormats(t *testing.T) {
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    write('Hello*n');
}`)

	tests := []TestConfig{
		{
			Name:      "default_executable",
			Args:      []string{"-o", filepath.Join(tmpDir, "test_default"), testFile},
			WantExit:  0,
			CheckFile: true,
			FileExt:   "",
		},
		{
			Name:      "llvm_ir_output",
			Args:      []string{"--emit-llvm", "-o", filepath.Join(tmpDir, "test.ll"), testFile},
			WantExit:  0,
			CheckFile: true,
			FileExt:   ".ll",
		},
		{
			Name:      "object_file_output",
			Args:      []string{"-c", "-o", filepath.Join(tmpDir, "test.o"), testFile},
			WantExit:  0,
			CheckFile: true,
			FileExt:   ".o",
		},
		{
			Name:      "assembly_output",
			Args:      []string{"-S", "-o", filepath.Join(tmpDir, "test.s"), testFile},
			WantExit:  0,
			CheckFile: true,
			FileExt:   ".s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}

// TestCLIOptimizationFlags tests optimization level flags
func TestCLIOptimizationFlags(t *testing.T) {
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    auto x;
    x = 42;
    return(x);
}`)

	tests := []TestConfig{
		{
			Name:     "optimization_O0",
			Args:     []string{"-O0", "-o", filepath.Join(tmpDir, "test_O0"), testFile},
			WantExit: 0,
		},
		{
			Name:     "optimization_O1",
			Args:     []string{"-O1", "-o", filepath.Join(tmpDir, "test_O1"), testFile},
			WantExit: 0,
		},
		{
			Name:     "optimization_O2",
			Args:     []string{"-O2", "-o", filepath.Join(tmpDir, "test_O2"), testFile},
			WantExit: 0,
		},
		{
			Name:     "optimization_O3",
			Args:     []string{"-O3", "-o", filepath.Join(tmpDir, "test_O3"), testFile},
			WantExit: 0,
		},
		{
			Name:     "invalid_optimization",
			Args:     []string{"-O4", testFile},
			WantExit: 1, // pflag correctly detects invalid optimization level
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}

// TestCLIDebugAndVerbose tests debug and verbose flags
func TestCLIDebugAndVerbose(t *testing.T) {
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    write('Test*n');
}`)

	tests := []TestConfig{
		{
			Name:       "verbose_output",
			Args:       []string{"-v", "-o", filepath.Join(tmpDir, "test_verbose"), testFile},
			WantExit:   0,
			WantOutput: "blang: compiling",
		},
		{
			Name:       "debug_info",
			Args:       []string{"-g", "-o", filepath.Join(tmpDir, "test_debug"), testFile},
			WantExit:   0,
			WantOutput: "",
		},
		{
			Name:       "verbose_and_debug",
			Args:       []string{"-v", "-g", "-o", filepath.Join(tmpDir, "test_verbose_debug"), testFile},
			WantExit:   0,
			WantOutput: "blang: compiling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}

// TestCLIPathFlags tests include and library path flags
func TestCLIPathFlags(t *testing.T) {
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    write('Test*n');
}`)

	tests := []TestConfig{
		{
			Name:     "library_path",
			Args:     []string{"-L", "/tmp", "-o", filepath.Join(tmpDir, "test_libpath"), testFile},
			WantExit: 0,
		},
		{
			Name:     "library_link",
			Args:     []string{"-l", "c", "-o", filepath.Join(tmpDir, "test_lib"), testFile},
			WantExit: 0,
		},
		{
			Name:     "multiple_library_dirs",
			Args:     []string{"-L", "/usr/lib", "-L", "/usr/local/lib", "-o", filepath.Join(tmpDir, "test_multi_libpath"), testFile},
			WantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}

// TestCLIStandardFlag tests the language standard flag
func TestCLIStandardFlag(t *testing.T) {
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    write('Test*n');
}`)

	tests := []TestConfig{
		{
			Name:     "default_standard",
			Args:     []string{"-o", filepath.Join(tmpDir, "test_default_std"), testFile},
			WantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}

// TestCLISaveTemps tests the save-temps flag
func TestCLISaveTemps(t *testing.T) {
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    write('Test*n');
}`)

	tests := []TestConfig{
		{
			Name:           "without_save_temps",
			Args:           []string{"-o", filepath.Join(tmpDir, "test_no_save"), testFile},
			WantExit:       0,
			ExpectTempFile: false,
		},
		{
			Name:           "with_save_temps",
			Args:           []string{"--save-temps", "-o", filepath.Join(tmpDir, "test_save"), testFile},
			WantExit:       0,
			ExpectTempFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}

// TestCLIExecutableGeneration tests that generated executables actually work
func TestCLIExecutableGeneration(t *testing.T) {
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    write('Hello*n');
}`)

	outputFile := filepath.Join(tmpDir, "test_executable")

	// Compile the executable
	cmd := exec.Command("./blang", "-o", outputFile, testFile)
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
	requireLibbO(t)

	tmpDir := t.TempDir()
	testFile := createTestFile(t, tmpDir, "test.b", `main() {
    write('Test*n');
}`)

	tests := []TestConfig{
		{
			Name:     "verbose_optimized_debug",
			Args:     []string{"-v", "-O2", "-g", "-o", filepath.Join(tmpDir, "test_combined1"), testFile},
			WantExit: 0,
		},
		{
			Name:     "verbose_optimized_debug_O1",
			Args:     []string{"-v", "-O1", "-g", "-o", filepath.Join(tmpDir, "test_combined2"), testFile},
			WantExit: 0,
		},
		{
			Name:           "all_flags",
			Args:           []string{"-v", "-O3", "-g", "--save-temps", "-o", filepath.Join(tmpDir, "test_combined3"), testFile},
			WantExit:       0,
			ExpectTempFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runBlangTest(t, tt)
		})
	}
}
