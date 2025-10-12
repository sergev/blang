package main

import (
	"fmt"
	"io"
	"os"
)

const (
	// ANSI color codes
	ColorReset     = "\033[0m"
	ColorBoldRed   = "\033[1m\033[31m"
	ColorBoldWhite = "\033[1m\033[37m"
)

// MaxFnCallArgs is the maximum number of function call arguments supported
const MaxFnCallArgs = 6

// ArgRegisters for x86_64 function calls
var ArgRegisters = []string{
	"%rdi",
	"%rsi",
	"%rdx",
	"%rcx",
	"%r8",
	"%r9",
}

// CompilerArgs holds the compiler state
type CompilerArgs struct {
	Arg0           string   // name of the executable
	OutputFile     string   // output file
	InputFiles     []string // input files
	NumInputFiles  int      // number of input files
	WordSize       int      // size of the B data type
	SaveTemps      bool     // should temporary files get deleted?
	Locals         *List    // local variables
	StackOffset    uint64   // local variable offset
	Extrns         *List    // extrn variables
	Strings        *List    // string table
	ConditionalCnt uint64   // counter for conditional labels
	StmtCnt        uint64   // counter for statement labels
}

// StackVar represents a stack variable
type StackVar struct {
	Name   string
	Offset uint64
}

// NewCompilerArgs creates a new CompilerArgs with default values
func NewCompilerArgs(arg0 string, inputFiles []string) *CompilerArgs {
	return &CompilerArgs{
		Arg0:       arg0,
		OutputFile: "a.s",
		InputFiles: inputFiles,
		WordSize:   8, // x86_64 architecture
		Locals:     NewList(),
		Extrns:     NewList(),
		Strings:    NewList(),
	}
}

// Eprintf prints an error message with prefix
func Eprintf(arg0 string, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s%s: %serror: %s", ColorBoldWhite, arg0, ColorBoldRed, ColorReset)
	fmt.Fprintf(os.Stderr, format, args...)
}

// FindIdentifier searches for an identifier in locals and extrns
// Returns offset and whether it's an extern
func (args *CompilerArgs) FindIdentifier(name string) (int64, bool, bool) {
	// Search in locals
	for i := 0; i < args.Locals.Size; i++ {
		v := args.Locals.Data[i].(*StackVar)
		if v.Name == name {
			return int64(v.Offset), false, true
		}
	}

	// Search in externs
	for i := 0; i < args.Extrns.Size; i++ {
		if args.Extrns.Data[i].(string) == name {
			return int64(i), true, true
		}
	}

	return -1, false, false
}

// Compile processes the input files and generates assembly
func Compile(args *CompilerArgs) error {
	// Create a buffer for the assembly code
	var buffer []byte

	// Open every provided `.b` file and generate assembly for it
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
		output, err := ParseDeclarations(lexer)
		if err != nil {
			file.Close()
			return err
		}
		buffer = append(buffer, output...)
		file.Close()
	}

	// Write the buffer to an assembly file
	outFile, err := os.Create(args.OutputFile)
	if err != nil {
		Eprintf(args.Arg0, "cannot open file '%s' %s.", args.OutputFile, err)
		return err
	}
	defer outFile.Close()

	_, err = outFile.Write(buffer)
	return err
}

// CmpOperator represents comparison operators
type CmpOperator int

const (
	CmpLT CmpOperator = iota // less-than
	CmpLE                    // less-than-equal
	CmpGT                    // greater-than
	CmpGE                    // greater-than-equal
	CmpEQ                    // equality
	CmpNE                    // non-equality
)

// CmpInstruction maps comparison operators to x86_64 set instructions
var CmpInstruction = []string{
	"setl",
	"setle",
	"setg",
	"setge",
	"sete",
	"setne",
}

// BinaryOperator represents binary operators
type BinaryOperator int

const (
	BinAdd BinaryOperator = iota // +
	BinSub                       // -
	BinMul                       // *
	BinDiv                       // /
	BinMod                       // %
	BinShl                       // <<
	BinSar                       // >>
	BinAnd                       // &
	BinOr                        // |
)

// BinaryCode maps binary operators to x86_64 assembly code
var BinaryCode = []string{
	// +
	"  pop %rdi\n" +
		"  add %rdi, %rax\n",

	// -
	"  mov %rax, %rdi\n" +
		"  pop %rax\n" +
		"  sub %rdi, %rax\n",

	// *
	"  pop %rdi\n" +
		"  imul %rdi, %rax\n",

	// /
	"  mov %rax, %rdi\n" +
		"  pop %rax\n" +
		"  cqo\n" +
		"  idiv %rdi\n",

	// %
	"  mov %rax, %rdi\n" +
		"  pop %rax\n" +
		"  cqo\n" +
		"  idiv %rdi\n" +
		"  mov %rdx, %rax\n",

	// <<
	"  mov %rax, %rcx\n" +
		"  pop %rax\n" +
		"  shl %cl, %rax\n",

	// >>
	"  mov %rax, %rcx\n" +
		"  pop %rax\n" +
		"  sar %cl, %rax\n",

	// &
	"  pop %rdi\n" +
		"  and %rdi, %rax\n",

	// |
	"  pop %rdi\n" +
		"  or %rdi, %rax\n",
}

// Writer interface for code generation
type Writer interface {
	io.Writer
	WriteString(s string) (int, error)
}
