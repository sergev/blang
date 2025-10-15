package main

import (
	"fmt"
	"os"
	"os/exec"

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
		OutputFile: "a.out", // default executable name
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
	// Create the compiler structure
	compiler := NewCompiler(args)

	// Open every provided `.b` file and generate LLVM IR for it
	for _, inputFile := range args.InputFiles {
		if len(inputFile) < 2 || inputFile[len(inputFile)-2:] != ".b" {
			continue
		}

		if args.Verbose {
			fmt.Printf("blang: processing %s\n", inputFile)
		}

		file, err := os.Open(inputFile)
		if err != nil {
			Eprintf(args.Arg0, "%s: %s\ncompilation terminated.\n", inputFile, err)
			return err
		}

		lexer := NewLexer(args, file)
		err = ParseDeclarations(lexer, compiler)
		if err != nil {
			file.Close()
			return err
		}
		file.Close()
	}

	// Write the LLVM IR to output file
	outFile, err := os.Create(args.OutputFile)
	if err != nil {
		Eprintf(args.Arg0, "cannot open file '%s' %s.", args.OutputFile, err)
		return err
	}
	defer outFile.Close()

	_, err = outFile.WriteString(compiler.GetModule().String())
	if err != nil {
		return err
	}

	if args.Verbose {
		fmt.Printf("blang: generated %s\n", args.OutputFile)
	}
	return nil
}

// compileToAssembly generates assembly output
func compileToAssembly(args *CompileOptions) error {
	// First generate LLVM IR
	tempIR := args.OutputFile + ".tmp.ll"
	originalOutput := args.OutputFile
	args.OutputFile = tempIR
	args.OutputType = OutputIR

	err := compileToIR(args)
	if err != nil {
		return err
	}

	// Convert LLVM IR to assembly using clang
	cmd := exec.Command("clang", "-S", "-o", originalOutput, tempIR)
	if args.Verbose {
		fmt.Printf("blang: running %s\n", cmd.String())
	}

	err = cmd.Run()
	if err != nil {
		os.Remove(tempIR)
		return fmt.Errorf("failed to generate assembly: %v", err)
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
	cmdArgs := []string{tempIR, "libb.o", "-o", originalOutput}

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
