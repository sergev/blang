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
	os.Exit(0)
}

func main() {
	var output string
	var saveTemps bool
	var showVersion bool
	var showHelp bool

	flag.StringVar(&output, "o", "", "place the output into <file>")
	flag.BoolVar(&saveTemps, "save-temps", false, "do not delete intermediate files")
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
	args := NewCompilerArgs("blang", files)
	args.SaveTemps = saveTemps
	if output != "" {
		args.OutputFile = output
	}

	// Compile
	if err := Compile(args); err != nil {
		Eprintf("blang", "%s\n", err)
		os.Exit(1)
	}
}
