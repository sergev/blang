package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// compile processes the input files and compiles them.
// This is a placeholder function to be developed later.
func compile(files []string, output string, verbose bool) error {
    // TODO: Implement compilation logic here.
    // For now, just print what would happen for demonstration.
    if verbose {
        fmt.Printf("Compiling files: %v\n", files)
        if output != "" {
            fmt.Printf("Output file: %s\n", output)
        } else {
            fmt.Println("Using default output: a.out")
        }
    }
    return nil
}

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
    var verbose bool
    var showVersion bool
    var showHelp bool

    flag.StringVar(&output, "o", "", "output file (default: a.out if multiple files, or derived from input)")
    flag.BoolVar(&verbose, "v", false, "enable verbose output")
    flag.BoolVar(&verbose, "verbose", false, "enable verbose output")
    flag.BoolVar(&showVersion, "version", false, "print version and exit")
    flag.BoolVar(&showHelp, "help", false, "print this help message and exit")

    flag.Usage = usage
    flag.Parse()

    if showHelp {
        usage()
    }

    if showVersion {
        fmt.Println("blang version 0.1")
        os.Exit(0)
    }

    files := flag.Args()
    if len(files) == 0 {
        fmt.Fprintf(os.Stderr, "Error: no input files\n")
        usage()
    }

    // Validate input files end with .b
    for _, file := range files {
        if !strings.HasSuffix(strings.ToLower(file), ".b") {
            fmt.Fprintf(os.Stderr, "Error: input file '%s' does not have .b extension\n", file)
            os.Exit(1)
        }
        if _, err := os.Stat(file); err != nil {
            fmt.Fprintf(os.Stderr, "Error: cannot access file '%s': %v\n", file, err)
            os.Exit(1)
        }
    }

    // Handle default output if not specified
    if output == "" {
        if len(files) == 1 {
            // Derive from single input file, e.g., foo.b -> foo
            base := filepath.Base(files[0])
            ext := filepath.Ext(base)
            output = base[:len(base)-len(ext)]
        } else {
            output = "a.out"
        }
    }

    // Process files in sequence
    if err := compile(files, output, verbose); err != nil {
        fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
        os.Exit(1)
    }

    if verbose {
        fmt.Println("Compilation successful")
    }
}
