package main

import (
	"fmt"
	"os"

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
	Arg0         string     // name of the executable
	OutputFile   string     // output file
	InputFiles   []string   // input files
	WordSize     int        // size of the B data type (8 for x86_64)
	SaveTemps    bool       // should temporary files get deleted?
	OutputType   OutputType // type of output to generate
	Optimize     int        // optimization level (0-3)
	DebugInfo    bool       // include debug information
	Verbose      bool       // verbose output
	LibraryDirs  []string   // library search directories
	Libraries    []string   // libraries to link
	GlobalPrefix string     // prefix for global symbols to avoid C clashes
}

// NewCompileOptions creates a new structure with default values
func NewCompileOptions(arg0 string, inputFiles []string) *CompileOptions {
	return &CompileOptions{
		Arg0:         arg0,
		InputFiles:   inputFiles,
		WordSize:     8, // x86_64 word size
		OutputType:   OutputExecutable,
		Optimize:     1, // optimization level -O1 by default
		GlobalPrefix: "b.",
	}
}

// Eprintf prints an error message with prefix
func Eprintf(arg0 string, format string, args ...interface{}) {
	color.New(color.FgWhite, color.Bold).Fprintf(os.Stderr, "%s: ", arg0)
	color.New(color.FgRed, color.Bold).Fprintf(os.Stderr, "error: ")
	fmt.Fprintf(os.Stderr, format, args...)
}
