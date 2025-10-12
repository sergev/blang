package main

import (
	"fmt"
	"os"
)

const (
	// ANSI color codes
	ColorReset     = "\033[0m"
	ColorBoldRed   = "\033[1m\033[31m"
	ColorBoldWhite = "\033[1m\033[37m"
)

// CompilerArgs holds the compiler state
type CompilerArgs struct {
	Arg0       string   // name of the executable
	OutputFile string   // output file
	InputFiles []string // input files
	WordSize   int      // size of the B data type (8 for x86_64)
	SaveTemps  bool     // should temporary files get deleted?
}

// NewCompilerArgs creates a new CompilerArgs with default values
func NewCompilerArgs(arg0 string, inputFiles []string) *CompilerArgs {
	return &CompilerArgs{
		Arg0:       arg0,
		OutputFile: "a.ll",
		InputFiles: inputFiles,
		WordSize:   8, // x86_64 word size
	}
}

// Eprintf prints an error message with prefix
func Eprintf(arg0 string, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s%s: %serror: %s", ColorBoldWhite, arg0, ColorBoldRed, ColorReset)
	fmt.Fprintf(os.Stderr, format, args...)
}

// Compile processes the input files and generates LLVM IR
func Compile(args *CompilerArgs) error {
	// Create LLVM compiler
	llvmCompiler := NewLLVMCompiler(args)

	// Open every provided `.b` file and generate LLVM IR for it
	for _, inputFile := range args.InputFiles {
		if len(inputFile) < 2 || inputFile[len(inputFile)-2:] != ".b" {
			continue
		}

		file, err := os.Open(inputFile)
		if err != nil {
			Eprintf(args.Arg0, "%s: %s\ncompilation terminated.\n", inputFile, err)
			return err
		}

		lexer := NewLexer(args, file)
		err = ParseDeclarationsLLVM(lexer, llvmCompiler)
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

	_, err = outFile.WriteString(llvmCompiler.GetModule().String())
	return err
}
