package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: blang [options] file...

blang is a compiler for .b files.

Options:
`)
	pflag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  blang hello.b                       # Compile to executable 'hello'
  blang -c hello.b                    # Compile to object file 'hello.o'
  blang -S hello.b                    # Compile to assembly 'hello.s'
  blang -emit-llvm hello.b            # Output LLVM IR 'hello.ll'
  blang -O2 -g -o optimized hello.b   # Optimized with debug info
  blang hello.b -o output -O2         # Options can be placed after arguments
  blang -V                            # Show version information

`)
	os.Exit(0)
}

func main() {
	var output string
	var saveTemps bool
	var showVersion bool
	var showHelp bool

	// Output format flags
	var compileOnly bool
	var assemblyOnly bool
	var emitLLVM bool

	// Optimization and debug flags
	var optimize string
	var debugInfo bool
	var verbose bool

	// Path flags
	var libraryDirs []string
	var libraries []string

	// Output control
	pflag.StringVarP(&output, "output", "o", "", "place the output into <file>")
	pflag.BoolVar(&saveTemps, "save-temps", false, "do not delete intermediate files")
	pflag.BoolVar(&emitLLVM, "emit-llvm", false, "emit LLVM IR instead of executable")

	// Compilation stages
	pflag.BoolVarP(&compileOnly, "compile", "c", false, "compile and assemble, but do not link")
	pflag.BoolVarP(&assemblyOnly, "assemble", "S", false, "compile only; do not assemble or link")

	// Optimization and debugging
	pflag.StringVarP(&optimize, "optimize", "O", "0", "optimization level (0-3)")
	pflag.BoolVarP(&debugInfo, "debug", "g", false, "generate debug information")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Paths and libraries
	pflag.StringSliceVarP(&libraryDirs, "library-dir", "L", []string{}, "add directory to library search path")
	pflag.StringSliceVarP(&libraries, "library", "l", []string{}, "link with library")

	// Help and version
	pflag.BoolVarP(&showVersion, "version", "V", false, "display compiler version information")
	pflag.BoolVarP(&showHelp, "help", "h", false, "display this information")

	pflag.Usage = usage
	pflag.Parse()

	if showHelp {
		usage()
	}

	if showVersion {
		fmt.Println("blang version 0.1")
		fmt.Println("Copyright (c) 2025 Serge Vakulenko")
		fmt.Println("Freely distributed under the MIT License.")
		fmt.Println("There is NO warranty.")
		os.Exit(0)
	}

	files := pflag.Args()

	// Check if no arguments at all were provided
	if len(os.Args) == 1 {
		fmt.Println("Usage: blang [options] file...")
		os.Exit(1)
	}

	if len(files) == 0 {
		Eprintf("blang", "no input files\ncompilation terminated.\n")
		os.Exit(1)
	}

	// Determine output type based on flags
	var outputType OutputType
	if assemblyOnly {
		outputType = OutputAssembly
	} else if compileOnly {
		outputType = OutputObject
	} else if emitLLVM {
		outputType = OutputIR
	} else {
		outputType = OutputExecutable
	}

	// Parse optimization level
	optLevel := 0
	if optimize != "0" {
		// Handle -O value format
		switch optimize {
		case "1", "2", "3":
			optLevel = int(optimize[0] - '0')
		default:
			Eprintf("blang", "invalid optimization level: %s\n", optimize)
			os.Exit(1)
		}
	}

	// Validate input file extensions
	allowedExt := map[string]bool{".b": true, ".ll": true, ".s": true, ".o": true, ".a": true}
	for _, file := range files {
		ext := filepath.Ext(file)
		if !allowedExt[ext] {
			Eprintf("blang", "unsupported input file extension for '%s'; allowed: .b, .ll, .s, .o, .a\n", file)
			os.Exit(1)
		}
		if _, err := os.Stat(file); err != nil {
			Eprintf("blang", "cannot access file '%s': %v\n", file, err)
			os.Exit(1)
		}
	}

	// Create compiler args
	args := NewCompileOptions("blang", files)
	args.SaveTemps = saveTemps
	args.OutputType = outputType
	args.Optimize = optLevel
	args.DebugInfo = debugInfo
	args.Verbose = verbose

	// Helper: append path if it exists and is a directory
	addIfDir := func(dst *[]string, p string) {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			*dst = append(*dst, p)
		}
	}

	// Build default library search paths (only those that exist)
	var defaults []string
	if home := os.Getenv("HOME"); home != "" {
		addIfDir(&defaults, filepath.Join(home, ".local/lib"))
	}
	addIfDir(&defaults, "/opt/homebrew/lib")
	addIfDir(&defaults, "/opt/local/lib")
	addIfDir(&defaults, "/usr/local/lib")
	addIfDir(&defaults, "/usr/lib")

	// Set library directories: defaults first, then user-specified
	args.LibraryDirs = append(defaults, libraryDirs...)

	// Set libraries
	args.Libraries = libraries

	// Set output file when -o provided; otherwise leave empty
	args.OutputFile = output

	// Compile
	if err := Compile(args); err != nil {
		Eprintf("blang", "%s\n", err)
		os.Exit(1)
	}
}
