package main

import (
	"fmt"
	"io"
	"unicode"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/value"
)

// ParseDeclarationsLLVM parses top-level declarations and generates LLVM IR
func ParseDeclarationsLLVM(l *Lexer, c *LLVMCompiler) error {
	for {
		name, err := l.Identifier()
		if err != nil {
			return err
		}
		if name == "" {
			break
		}

		ch, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("unexpected end of file after declaration")
			}
			return err
		}

		switch ch {
		case '(':
			if err := parseFunctionLLVM(l, c, name); err != nil {
				return err
			}
		case '[':
			if err := parseVectorLLVM(l, c, name); err != nil {
				return err
			}
		default:
			l.UnreadChar(ch)
			if err := parseGlobalLLVM(l, c, name); err != nil {
				return err
			}
		}
	}

	// Check for unexpected input
	_, err := l.ReadChar()
	if err != io.EOF {
		if err == nil {
			return fmt.Errorf("expect identifier at top level")
		}
		return err
	}

	return nil
}

// parseGlobalLLVM parses a global variable
func parseGlobalLLVM(l *Lexer, c *LLVMCompiler, name string) error {
	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	if ch != ';' {
		l.UnreadChar(ch)
		// Parse initialization list
		init, err := parseIvalConstLLVM(l, c)
		if err != nil {
			return err
		}
		c.DeclareGlobal(name, init)

		if err := l.Whitespace(); err != nil {
			return err
		}
		ch, err = l.ReadChar()
		if err != nil {
			return err
		}
		if ch != ';' {
			return fmt.Errorf("expect ';' at end of declaration")
		}
	} else {
		c.DeclareGlobal(name, nil)
	}

	return nil
}

// parseVectorLLVM parses a global array
func parseVectorLLVM(l *Lexer, c *LLVMCompiler, name string) error {
	var nwords int64 = 0

	if err := l.Whitespace(); err != nil {
		return err
	}

	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	if ch != ']' {
		l.UnreadChar(ch)
		nwords, err = l.Number()
		if err != nil {
			return fmt.Errorf("unexpected end of file, expect vector size after '['")
		}

		if err := l.Whitespace(); err != nil {
			return err
		}

		if err := l.ExpectChar(']', "expect ']' after vector size"); err != nil {
			return err
		}
	}

	if err := l.Whitespace(); err != nil {
		return err
	}

	ch, err = l.ReadChar()
	if err != nil {
		return err
	}

	var initVals []constant.Constant
	if ch != ';' {
		l.UnreadChar(ch)
		for {
			if err := l.Whitespace(); err != nil {
				return err
			}
			val, err := parseIvalConstLLVM(l, c)
			if err != nil {
				return err
			}
			initVals = append(initVals, val)
			if err := l.Whitespace(); err != nil {
				return err
			}

			ch, err := l.ReadChar()
			if err != nil {
				return err
			}
			if ch == ';' {
				break
			}
			if ch != ',' {
				return fmt.Errorf("expect ';' at end of declaration")
			}
		}
	}

	if nwords == 0 {
		nwords = int64(len(initVals))
	}

	c.DeclareGlobalArray(name, nwords, initVals)
	return nil
}

// parseIvalConstLLVM parses a constant initialization value
func parseIvalConstLLVM(l *Lexer, c *LLVMCompiler) (constant.Constant, error) {
	ch, err := l.ReadChar()
	if err != nil {
		return nil, err
	}

	if unicode.IsLetter(ch) {
		l.UnreadChar(ch)
		name, err := l.Identifier()
		if err != nil || name == "" {
			return nil, fmt.Errorf("unexpected end of file, expect ival")
		}
		// For now, just return a zero constant for references
		// TODO: Handle proper global references
		return constant.NewInt(c.WordType(), 0), nil
	} else if ch == '\'' {
		val, err := l.Character()
		if err != nil {
			return nil, fmt.Errorf("unexpected end of file, expect ival")
		}
		return constant.NewInt(c.WordType(), val), nil
	} else if ch == '"' {
		str, err := l.String()
		if err != nil {
			return nil, err
		}
		global := c.CreateStringConstant(str)
		return global, nil
	} else if ch == '-' {
		val, err := l.Number()
		if err != nil {
			return nil, fmt.Errorf("unexpected end of file, expect ival")
		}
		return constant.NewInt(c.WordType(), -val), nil
	} else {
		l.UnreadChar(ch)
		val, err := l.Number()
		if err != nil {
			return nil, fmt.Errorf("unexpected end of file, expect ival")
		}
		return constant.NewInt(c.WordType(), val), nil
	}
}

// parseFunctionLLVM parses a function definition
func parseFunctionLLVM(l *Lexer, c *LLVMCompiler, name string) error {
	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	var paramNames []string
	if ch != ')' {
		l.UnreadChar(ch)
		paramNames, err = parseArgumentsLLVM(l)
		if err != nil {
			return err
		}
	}

	fn := c.DeclareFunction(name, paramNames)
	c.StartFunction(fn)

	if err := parseStatementLLVMWithSwitch(l, c, -1, nil); err != nil {
		return err
	}

	c.EndFunction()
	return nil
}

// parseArgumentsLLVM parses function arguments
func parseArgumentsLLVM(l *Lexer) ([]string, error) {
	var params []string

	for {
		if err := l.Whitespace(); err != nil {
			return nil, err
		}

		name, err := l.Identifier()
		if err != nil || name == "" {
			return nil, fmt.Errorf("expect ')' or identifier after function arguments")
		}

		params = append(params, name)

		if err := l.Whitespace(); err != nil {
			return nil, err
		}

		ch, err := l.ReadChar()
		if err != nil {
			return nil, err
		}

		switch ch {
		case ')':
			return params, nil
		case ',':
			continue
		default:
			return nil, fmt.Errorf("unexpected character '%c', expect ')' or ','", ch)
		}
	}
}

// parseStatementLLVM parses a statement
// switchID: ID of enclosing switch statement (-1 if none)
// cases: slice to accumulate case values for switch
func parseStatementLLVM(l *Lexer, c *LLVMCompiler) error {
	return parseStatementLLVMWithSwitch(l, c, -1, nil)
}

func parseStatementLLVMWithSwitch(l *Lexer, c *LLVMCompiler, switchID int64, cases *[]int64) error {
	if err := l.Whitespace(); err != nil {
		return err
	}

	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	switch ch {
	case '{':
		// Block statement
		for {
			if err := l.Whitespace(); err != nil {
				return err
			}
			ch, err := l.ReadChar()
			if err != nil {
				return err
			}
			if ch == '}' {
				break
			}
			l.UnreadChar(ch)
			if err := parseStatementLLVMWithSwitch(l, c, switchID, cases); err != nil {
				return err
			}
		}

	case ';':
		// Null statement

	default:
		if unicode.IsLetter(ch) {
			l.UnreadChar(ch)
			return parseKeywordOrExpressionLLVMWithSwitch(l, c, switchID, cases)
		} else {
			l.UnreadChar(ch)
			// Expression statement
			_, err := parseExpressionLLVM(l, c)
			if err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			if err := l.ExpectChar(';', "expect ';' after expression statement"); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseKeywordOrExpressionLLVM handles keywords and expressions
func parseKeywordOrExpressionLLVM(l *Lexer, c *LLVMCompiler) error {
	return parseKeywordOrExpressionLLVMWithSwitch(l, c, -1, nil)
}

func parseKeywordOrExpressionLLVMWithSwitch(l *Lexer, c *LLVMCompiler, switchID int64, cases *[]int64) error {
	name, err := l.Identifier()
	if err != nil {
		return err
	}

	if err := l.Whitespace(); err != nil {
		return err
	}

	switch name {
	case "return":
		return parseReturnLLVM(l, c)
	case "auto":
		return parseAutoLLVM(l, c)
	case "extrn":
		return parseExtrnLLVM(l, c)
	case "if":
		return parseIfLLVM(l, c)
	case "while":
		return parseWhileLLVM(l, c)
	case "switch":
		return parseSwitchLLVM(l, c)
	case "case":
		return parseCaseLLVM(l, c, switchID, cases)
	case "goto":
		return parseGotoLLVM(l, c)
	default:
		// Check if it's a label
		ch, err := l.ReadChar()
		if err != nil {
			return err
		}
		if ch == ':' {
			// Label - get or create labeled block
			block := c.GetOrCreateLabel(name)

			// Only create a branch if the current block has no terminator
			// This handles fall-through to labels
			currentBlock := c.GetInsertBlock()
			if currentBlock != nil && currentBlock.Term == nil {
				c.builder.NewBr(block)
			}

			// Set insertion point to the label block
			c.SetInsertPoint(block)
			return parseStatementLLVMWithSwitch(l, c, switchID, cases)
		}

		// Otherwise it's an expression
		l.UnreadChar(ch)
		for i := len(name) - 1; i >= 0; i-- {
			l.UnreadChar(rune(name[i]))
		}
		_, err = parseExpressionLLVM(l, c)
		if err != nil {
			return err
		}
		if err := l.Whitespace(); err != nil {
			return err
		}
		if err := l.ExpectChar(';', "expect ';' after expression statement"); err != nil {
			return err
		}
	}

	return nil
}

// parseReturnLLVM parses a return statement
func parseReturnLLVM(l *Lexer, c *LLVMCompiler) error {
	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	if ch != ';' {
		if ch != '(' {
			return fmt.Errorf("expect '(' or ';' after 'return'")
		}
		val, err := parseExpressionLLVM(l, c)
		if err != nil {
			return err
		}
		if err := l.Whitespace(); err != nil {
			return err
		}
		if err := l.ExpectChar(')', "expect ')' after 'return' statement"); err != nil {
			return err
		}
		if err := l.Whitespace(); err != nil {
			return err
		}
		if err := l.ExpectChar(';', "expect ';' after 'return' statement"); err != nil {
			return err
		}
		c.builder.NewRet(val)
	} else {
		c.builder.NewRet(constant.NewInt(c.WordType(), 0))
	}

	return nil
}

// parseAutoLLVM parses auto variable declarations
func parseAutoLLVM(l *Lexer, c *LLVMCompiler) error {
	for {
		name, err := l.Identifier()
		if err != nil || name == "" {
			return fmt.Errorf("expect identifier after 'auto'")
		}

		if err := l.Whitespace(); err != nil {
			return err
		}

		ch, err := l.ReadChar()
		if err != nil {
			return err
		}

		if ch == '[' {
			// Array declaration
			size, err := l.Number()
			if err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			if err := l.ExpectChar(']', "expect ']' after array size"); err != nil {
				return err
			}
			c.DeclareLocalArray(name, size)

			if err := l.Whitespace(); err != nil {
				return err
			}
			ch, err = l.ReadChar()
			if err != nil {
				return err
			}
		} else {
			// Scalar variable
			c.DeclareLocal(name)
		}

		if ch == ';' {
			break
		}
		if ch != ',' {
			return fmt.Errorf("unexpected character '%c', expect ';' or ','", ch)
		}
	}

	return nil
}

// parseExtrnLLVM parses external declarations
func parseExtrnLLVM(l *Lexer, c *LLVMCompiler) error {
	for {
		name, err := l.Identifier()
		if err != nil || name == "" {
			return fmt.Errorf("expect identifier after 'extrn'")
		}

		// In B, extrn can refer to global variables or functions
		// We don't know which at declaration time, so we declare both as potential
		// The usage will determine the actual type
		// For now, just remember we saw this extrn declaration
		// If it's used as a function later, it will be auto-declared

		if _, exists := c.globals[name]; !exists {
			if _, exists := c.functions[name]; !exists {
				// Declare as external global variable (i64)
				// If it turns out to be a function, it will be redeclared later
				global := c.module.NewGlobal(name, c.WordType())
				c.globals[name] = global
			}
		}

		if err := l.Whitespace(); err != nil {
			return err
		}

		ch, err := l.ReadChar()
		if err != nil {
			return err
		}

		if ch == ';' {
			return nil
		}
		if ch != ',' {
			return fmt.Errorf("unexpected character '%c', expect ';' or ','", ch)
		}
	}
}

// parseIfLLVM parses if statements
func parseIfLLVM(l *Lexer, c *LLVMCompiler) error {
	if err := l.ExpectChar('(', "expect '(' after 'if'"); err != nil {
		return err
	}

	cond, err := parseExpressionLLVM(l, c)
	if err != nil {
		return err
	}

	if err := l.Whitespace(); err != nil {
		return err
	}
	if err := l.ExpectChar(')', "expect ')' after condition"); err != nil {
		return err
	}

	// Create blocks
	thenBlock := c.NewBlock("if.then")
	elseBlock := c.NewBlock("if.else")
	endBlock := c.NewBlock("if.end")

	// Compare condition to zero
	zero := constant.NewInt(c.WordType(), 0)
	cmp := c.builder.NewICmp(enum.IPredNE, cond, zero)
	c.builder.NewCondBr(cmp, thenBlock, elseBlock)

	// Generate then block
	c.SetInsertPoint(thenBlock)
	if err := parseStatementLLVM(l, c); err != nil {
		return err
	}
	if c.builder.Term == nil {
		c.builder.NewBr(endBlock)
	}

	// Check for else
	c.SetInsertPoint(elseBlock)
	if err := l.Whitespace(); err != nil {
		return err
	}

	// Try to read "else"
	elseChars := []rune{'e', 'l', 's', 'e'}
	var readChars []rune
	isElse := true

	for _, expected := range elseChars {
		ch, err := l.ReadChar()
		if err != nil || ch != expected {
			isElse = false
			if err == nil {
				readChars = append(readChars, ch)
			}
			break
		}
		readChars = append(readChars, ch)
	}

	if isElse {
		// Check that next char is not alphanumeric
		ch, err := l.ReadChar()
		if err == nil {
			readChars = append(readChars, ch)
			if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
				isElse = false
			}
		}
	}

	if isElse {
		if err := parseStatementLLVM(l, c); err != nil {
			return err
		}
	} else {
		// Push back characters
		for i := len(readChars) - 1; i >= 0; i-- {
			l.UnreadChar(readChars[i])
		}
	}

	if c.builder.Term == nil {
		c.builder.NewBr(endBlock)
	}

	c.SetInsertPoint(endBlock)
	return nil
}

// parseWhileLLVM parses while loops
func parseWhileLLVM(l *Lexer, c *LLVMCompiler) error {
	if err := l.ExpectChar('(', "expect '(' after 'while'"); err != nil {
		return err
	}

	// Create blocks
	condBlock := c.NewBlock("while.cond")
	bodyBlock := c.NewBlock("while.body")
	endBlock := c.NewBlock("while.end")

	// Jump to condition
	c.builder.NewBr(condBlock)
	c.SetInsertPoint(condBlock)

	// Evaluate condition
	cond, err := parseExpressionLLVM(l, c)
	if err != nil {
		return err
	}

	if err := l.Whitespace(); err != nil {
		return err
	}
	if err := l.ExpectChar(')', "expect ')' after condition"); err != nil {
		return err
	}

	// Compare condition to zero
	zero := constant.NewInt(c.WordType(), 0)
	cmp := c.builder.NewICmp(enum.IPredNE, cond, zero)
	c.builder.NewCondBr(cmp, bodyBlock, endBlock)

	// Generate body
	c.SetInsertPoint(bodyBlock)
	if err := parseStatementLLVM(l, c); err != nil {
		return err
	}
	if c.builder.Term == nil {
		c.builder.NewBr(condBlock)
	}

	c.SetInsertPoint(endBlock)
	return nil
}

// parseExpressionLLVM parses an expression and returns the result value
// This is a wrapper that calls the comprehensive expression parser with full precedence support
func parseExpressionLLVM(l *Lexer, c *LLVMCompiler) (value.Value, error) {
	return parseExpressionLLVMWithLevel(l, c, 15)
}
