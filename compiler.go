package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// OutputType represents different output formats
type OutputType int

const (
	OutputExecutable OutputType = iota // default - executable
	OutputObject                       // -c: object file
	OutputAssembly                     // -S: assembly file
	OutputIR                           // -emit-llvm: LLVM IR
)

// CompileOptions holds the compiler state
type CompileOptions struct {
	Arg0        string     // name of the executable
	OutputFile  string     // output file
	InputFiles  []string   // input files
	WordSize    int        // size of the B data type (8 for x86_64)
	SaveTemps   bool       // should temporary files get deleted?
	OutputType  OutputType // type of output to generate
	Optimize    int        // optimization level (0-3)
	DebugInfo   bool       // include debug information
	Verbose     bool       // verbose output
	LibraryDirs []string   // library search directories
	Libraries   []string   // libraries to link
}

// NewCompileOptions creates a new structure with default values
func NewCompileOptions(arg0 string, inputFiles []string) *CompileOptions {
	return &CompileOptions{
		Arg0:       arg0,
		InputFiles: inputFiles,
		WordSize:   8, // x86_64 word size
		OutputType: OutputExecutable,
		Optimize:   0, // no optimization by default
	}
}

// Eprintf prints an error message with prefix
func Eprintf(arg0 string, format string, args ...interface{}) {
	color.New(color.FgWhite, color.Bold).Fprintf(os.Stderr, "%s: ", arg0)
	color.New(color.FgRed, color.Bold).Fprintf(os.Stderr, "error: ")
	fmt.Fprintf(os.Stderr, format, args...)
}

// Compile processes the input files and generates the requested output format
func Compile(args *CompileOptions) error {
	if args.Verbose {
		fmt.Printf("blang: compiling %d file(s)\n", len(args.InputFiles))
	}

	// Handle different output types
	switch args.OutputType {
	case OutputIR:
		return compileToIR(args)
	case OutputAssembly:
		return compileToAssembly(args)
	case OutputObject:
		return compileToObject(args)
	case OutputExecutable:
		return compileToExecutable(args)
	default:
		return fmt.Errorf("unsupported output type")
	}
}

// compileToIR generates LLVM IR output
func compileToIR(args *CompileOptions) error {
	// Helper to compile a single .b file to the provided output path
	compileSingleTo := func(inputFile string, outputPath string) error {
		if args.Verbose {
			fmt.Printf("blang: processing %s\n", inputFile)
		}
		file, err := os.Open(inputFile)
		if err != nil {
			Eprintf(args.Arg0, "%s: %s\ncompilation terminated.\n", inputFile, err)
			return err
		}
		defer file.Close()

		// Create a fresh compiler per output unit
		compiler := NewCompiler(args)
		lexer := NewLexer(args, file)
		if err := ParseDeclarations(lexer, compiler); err != nil {
			return err
		}

		outFile, err := os.Create(outputPath)
		if err != nil {
			Eprintf(args.Arg0, "cannot open file '%s' %s.", outputPath, err)
			return err
		}
		if _, err := outFile.WriteString(compiler.GetModule().String()); err != nil {
			outFile.Close()
			return err
		}
		if err := outFile.Close(); err != nil {
			return err
		}
		if args.Verbose {
			fmt.Printf("blang: generated %s\n", outputPath)
		}
		return nil
	}

	// In non-pipeline IR mode, enforce that all inputs are .b files
	for _, inputFile := range args.InputFiles {
		if !strings.HasSuffix(inputFile, ".b") {
			return fmt.Errorf("input file '%s' does not have .b extension", inputFile)
		}
	}

	if args.OutputFile != "" {
		// -o present: exactly one input file
		if len(args.InputFiles) != 1 {
			return fmt.Errorf("multiple input files with -o for IR output are not allowed")
		}
		return compileSingleTo(args.InputFiles[0], args.OutputFile)
	}

	// No -o: emit one .ll per input into current working directory
	for _, inputFile := range args.InputFiles {
		base := filepath.Base(inputFile)
		out := strings.TrimSuffix(base, filepath.Ext(base)) + ".ll"
		if err := compileSingleTo(inputFile, out); err != nil {
			return err
		}
	}
	return nil
}

// compileToAssembly generates assembly output
func compileToAssembly(args *CompileOptions) error {
	// Validate extensions: only .b and .ll are accepted
	for _, in := range args.InputFiles {
		ext := filepath.Ext(in)
		if ext != ".b" && ext != ".ll" {
			return fmt.Errorf("input file '%s' must have .b or .ll extension", in)
		}
	}

	// Helper to build clang args including optimization/debug flags
	buildClangArgs := func(inputIRorLL, out string) []string {
		cmdArgs := []string{}
		if args.Optimize > 0 {
			cmdArgs = append(cmdArgs, fmt.Sprintf("-O%d", args.Optimize))
		}
		if args.DebugInfo {
			cmdArgs = append(cmdArgs, "-g")
		}
		cmdArgs = append(cmdArgs, "-S", "-o", out, inputIRorLL)
		return cmdArgs
	}

	// Process a single input into the specified output path
	processOne := func(in, out string) error {
		if strings.HasSuffix(in, ".b") {
			// Compile .b to temporary IR first
			tempIR := out + ".tmp.ll"

			irArgs := *args
			irArgs.InputFiles = []string{in}
			irArgs.OutputType = OutputIR
			irArgs.OutputFile = tempIR

			if err := compileToIR(&irArgs); err != nil {
				return err
			}

			// Convert IR to assembly using clang
			cmd := exec.Command("clang", buildClangArgs(tempIR, out)...)
			if args.Verbose {
				fmt.Printf("blang: running %s\n", cmd.String())
			}
			if err := cmd.Run(); err != nil {
				if !args.SaveTemps {
					os.Remove(tempIR)
				}
				return fmt.Errorf("failed to generate assembly: %v", err)
			}

			// Clean up temporary IR unless save-temps is specified
			if !args.SaveTemps {
				os.Remove(tempIR)
			}

			if args.Verbose {
				fmt.Printf("blang: generated %s\n", out)
			}
			return nil
		}

		// .ll: directly assemble with clang
		cmd := exec.Command("clang", buildClangArgs(in, out)...)
		if args.Verbose {
			fmt.Printf("blang: running %s\n", cmd.String())
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to generate assembly: %v", err)
		}
		if args.Verbose {
			fmt.Printf("blang: generated %s\n", out)
		}
		return nil
	}

	// If -o specified, require exactly one input file
	if args.OutputFile != "" {
		if len(args.InputFiles) != 1 {
			return fmt.Errorf("-o requires exactly one input file")
		}
		return processOne(args.InputFiles[0], args.OutputFile)
	}

	// No -o: emit one .s per input in the current working directory
	for _, in := range args.InputFiles {
		base := filepath.Base(in)
		out := strings.TrimSuffix(base, filepath.Ext(base)) + ".s"
		if err := processOne(in, out); err != nil {
			return err
		}
	}
	return nil
}

// compileToObject generates object file
func compileToObject(args *CompileOptions) error {
	// First generate LLVM IR
	tempIR := args.OutputFile + ".tmp.ll"
	originalOutput := args.OutputFile
	args.OutputFile = tempIR
	args.OutputType = OutputIR

	err := compileToIR(args)
	if err != nil {
		return err
	}

	// Convert LLVM IR to object file using clang
	cmd := exec.Command("clang", "-c", "-o", originalOutput, tempIR)
	if args.Verbose {
		fmt.Printf("blang: running %s\n", cmd.String())
	}

	err = cmd.Run()
	if err != nil {
		os.Remove(tempIR)
		return fmt.Errorf("failed to generate object file: %v", err)
	}

	// Clean up temporary file unless save-temps is specified
	if !args.SaveTemps {
		os.Remove(tempIR)
	}

	if args.Verbose {
		fmt.Printf("blang: generated %s\n", originalOutput)
	}
	return nil
}

// compileToExecutable generates executable
func compileToExecutable(args *CompileOptions) error {
	// First generate LLVM IR
	tempIR := args.OutputFile + ".tmp.ll"
	originalOutput := args.OutputFile
	args.OutputFile = tempIR
	args.OutputType = OutputIR

	err := compileToIR(args)
	if err != nil {
		return err
	}

	// Build clang command for linking
	cmdArgs := []string{tempIR, "-Lruntime", "-lb", "-o", originalOutput}

	// Add optimization flags if specified
	if args.Optimize > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-O%d", args.Optimize))
	}

	// Add debug info if requested
	if args.DebugInfo {
		cmdArgs = append(cmdArgs, "-g")
	}

	// Add library directories
	for _, libDir := range args.LibraryDirs {
		cmdArgs = append(cmdArgs, "-L"+libDir)
	}

	// Add libraries
	for _, lib := range args.Libraries {
		cmdArgs = append(cmdArgs, "-l"+lib)
	}

	cmd := exec.Command("clang", cmdArgs...)
	if args.Verbose {
		fmt.Printf("blang: running %s\n", cmd.String())
	}

	err = cmd.Run()
	if err != nil {
		os.Remove(tempIR)
		return fmt.Errorf("failed to generate executable: %v", err)
	}

	// Clean up temporary file unless save-temps is specified
	if !args.SaveTemps {
		os.Remove(tempIR)
	}

	if args.Verbose {
		fmt.Printf("blang: generated %s\n", originalOutput)
	}
	return nil
}
