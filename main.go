package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: blang [options] file...

blang is a compiler for .b files.

Options:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  blang -o hello hello.b              # Compile to executable
  blang -c hello.b                    # Compile to object file
  blang -S hello.b                    # Compile to assembly
  blang -emit-llvm hello.b            # Output LLVM IR
  blang -O2 -g -o optimized hello.b   # Optimized with debug info
  blang -Wall -Wextra hello.b         # Enable warnings

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
	var preprocess bool

	// Optimization and debug flags
	var optimize string
	var debugInfo bool
	var verbose bool

	// Warning flags
	var warnings bool
	var warningsAsErrors bool

	// Path flags
	var includeDirs string
	var libraryDirs string
	var libraries string

	// Other flags
	var standard string

	// Output control
	flag.StringVar(&output, "o", "", "place the output into <file>")
	flag.BoolVar(&saveTemps, "save-temps", false, "do not delete intermediate files")
	flag.BoolVar(&emitLLVM, "emit-llvm", false, "emit LLVM IR instead of executable")

	// Compilation stages
	flag.BoolVar(&compileOnly, "c", false, "compile and assemble, but do not link")
	flag.BoolVar(&assemblyOnly, "S", false, "compile only; do not assemble or link")
	flag.BoolVar(&preprocess, "E", false, "preprocess only; do not compile, assemble or link")

	// Optimization and debugging
	flag.StringVar(&optimize, "O", "0", "optimization level (0-3)")
	flag.BoolVar(&debugInfo, "g", false, "generate debug information")
	flag.BoolVar(&verbose, "v", false, "verbose output")

	// Handle optimization flags like -O2, -O3, etc.
	flag.Bool("O0", false, "no optimization")
	flag.Bool("O1", false, "optimization level 1")
	flag.Bool("O2", false, "optimization level 2")
	flag.Bool("O3", false, "optimization level 3")

	// Warnings
	flag.BoolVar(&warnings, "Wall", false, "enable all warnings")
	flag.BoolVar(&warningsAsErrors, "Werror", false, "treat warnings as errors")

	// Paths and libraries
	flag.StringVar(&includeDirs, "I", "", "add directory to include search path (comma-separated)")
	flag.StringVar(&libraryDirs, "L", "", "add directory to library search path (comma-separated)")
	flag.StringVar(&libraries, "l", "", "link with library (comma-separated)")

	// Language standard
	flag.StringVar(&standard, "std", "b", "language standard to use")

	// Help and version
	flag.BoolVar(&showVersion, "version", false, "display compiler version information")
	flag.BoolVar(&showHelp, "help", false, "display this information")

	flag.Usage = usage
	flag.Parse()

	if showHelp {
		usage()
	}

	if showVersion {
		fmt.Println("blang version 0.1")
		fmt.Println("Copyright (C) 2025")
		fmt.Println("This is free software; see the source for copying conditions.")
		fmt.Println("There is NO warranty.")
		os.Exit(0)
	}

	files := flag.Args()
	if len(files) == 0 {
		Eprintf("blang", "no input files\ncompilation terminated.\n")
		os.Exit(1)
	}

	// Determine output type based on flags
	var outputType OutputType
	if preprocess {
		outputType = OutputPreprocessed
	} else if assemblyOnly {
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

	// Check for optimization flags set via flag package
	if flag.Lookup("O0").Value.(flag.Getter).Get().(bool) {
		optLevel = 0
	} else if flag.Lookup("O1").Value.(flag.Getter).Get().(bool) {
		optLevel = 1
	} else if flag.Lookup("O2").Value.(flag.Getter).Get().(bool) {
		optLevel = 2
	} else if flag.Lookup("O3").Value.(flag.Getter).Get().(bool) {
		optLevel = 3
	} else if optimize != "0" {
		// Handle -O value format
		switch optimize {
		case "1", "2", "3":
			optLevel = int(optimize[0] - '0')
		default:
			Eprintf("blang", "invalid optimization level: %s\n", optimize)
			os.Exit(1)
		}
	}

	// Validate input files end with .b
	for _, file := range files {
		if len(file) < 2 || !strings.HasSuffix(file, ".b") {
			Eprintf("blang", "input file '%s' does not have .b extension\n", file)
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
	args.Warnings = warnings
	args.WarningsAsErrors = warningsAsErrors
	args.Standard = standard

	// Parse include directories
	if includeDirs != "" {
		args.IncludeDirs = strings.Split(includeDirs, ",")
	}

	// Parse library directories
	if libraryDirs != "" {
		args.LibraryDirs = strings.Split(libraryDirs, ",")
	}

	// Parse libraries
	if libraries != "" {
		args.Libraries = strings.Split(libraries, ",")
	}

	// Set output file
	if output != "" {
		args.OutputFile = output
	} else {
		// Set default output based on output type
		switch outputType {
		case OutputObject:
			args.OutputFile = strings.TrimSuffix(files[0], ".b") + ".o"
		case OutputAssembly:
			args.OutputFile = strings.TrimSuffix(files[0], ".b") + ".s"
		case OutputIR:
			args.OutputFile = strings.TrimSuffix(files[0], ".b") + ".ll"
		case OutputPreprocessed:
			args.OutputFile = strings.TrimSuffix(files[0], ".b") + ".i"
		default:
			args.OutputFile = "a.out"
		}
	}

	// Compile
	if err := Compile(args); err != nil {
		Eprintf("blang", "%s\n", err)
		os.Exit(1)
	}
}
