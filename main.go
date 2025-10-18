package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

func usage() {
	hdr := color.New(color.FgCyan, color.Bold)
	cmd := color.New(color.FgGreen, color.Bold)
	out := color.New(color.FgRed, color.Bold)
	note := color.New(color.Faint)

	hdr.Fprintln(os.Stderr, "Usage: blang [options] file...")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "blang is a compiler for .b files.")
	fmt.Fprintln(os.Stderr)

	hdr.Fprintln(os.Stderr, "Options:")
	pflag.PrintDefaults()
	fmt.Fprintln(os.Stderr)

	hdr.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintf(os.Stderr, "  %s  %s%s%s\n", cmd.Sprint("blang hello.b"), note.Sprint("                Compile to executable '"), out.Sprint("hello"), note.Sprint("'"))
	fmt.Fprintf(os.Stderr, "  %s  %s%s%s\n", cmd.Sprint("blang -c hello.b"), note.Sprint("             Compile to object file '"), out.Sprint("hello.o"), note.Sprint("'"))
	fmt.Fprintf(os.Stderr, "  %s  %s%s%s\n", cmd.Sprint("blang -S hello.b"), note.Sprint("             Compile to assembly '"), out.Sprint("hello.s"), note.Sprint("'"))
	fmt.Fprintf(os.Stderr, "  %s  %s%s%s\n", cmd.Sprint("blang -emit-llvm hello.b"), note.Sprint("     Output LLVM IR '"), out.Sprint("hello.ll"), note.Sprint("'"))
	fmt.Fprintf(os.Stderr, "  %s  %s\n", cmd.Sprint("blang -O0 -g -o unopt hello.b"), note.Sprint("Unoptimized with debug info"))
	fmt.Fprintf(os.Stderr, "  %s  %s\n", cmd.Sprint("blang hello.b -o output -O2"), note.Sprint("  Options can be placed after arguments"))
	fmt.Fprintf(os.Stderr, "  %s  %s\n", cmd.Sprint("blang -V"), note.Sprint("                     Show version information"))
	fmt.Fprintf(os.Stderr, "\n")
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
	pflag.StringVarP(&output, "output", "o", "", "Place the output into <file>")
	pflag.BoolVar(&saveTemps, "save-temps", false, "Do not delete intermediate files")
	pflag.BoolVar(&emitLLVM, "emit-llvm", false, "Emit LLVM IR instead of executable")

	// Compilation stages
	pflag.BoolVarP(&compileOnly, "compile", "c", false, "Compile and assemble, but do not link")
	pflag.BoolVarP(&assemblyOnly, "assemble", "S", false, "Compile only; do not assemble or link")

	// Optimization and debugging
	pflag.StringVarP(&optimize, "optimize", "O", "0", "Optimization level (0-3)")
	pflag.BoolVarP(&debugInfo, "debug", "g", false, "Generate debug information")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Paths and libraries
	pflag.StringSliceVarP(&libraryDirs, "library-dir", "L", []string{}, "Add directory to library search path")
	pflag.StringSliceVarP(&libraries, "library", "l", []string{}, "Link with library")

	// Help and version
	pflag.BoolVarP(&showVersion, "version", "V", false, "Display compiler version information")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Display this information")

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
		fmt.Println("Use '-h' to see all available options.")
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
