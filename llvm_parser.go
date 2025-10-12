package main

import (
	"fmt"
	"io"
	"unicode"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
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

	if err := parseStatementLLVM(l, c); err != nil {
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
func parseStatementLLVM(l *Lexer, c *LLVMCompiler) error {
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
			if err := parseStatementLLVM(l, c); err != nil {
				return err
			}
		}

	case ';':
		// Null statement

	default:
		if unicode.IsLetter(ch) {
			l.UnreadChar(ch)
			return parseKeywordOrExpressionLLVM(l, c)
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
	default:
		// Check if it's a label
		ch, err := l.ReadChar()
		if err != nil {
			return err
		}
		if ch == ':' {
			// Label - create a new block
			block := c.NewBlock(name)
			if c.builder.Term == nil {
				c.builder.NewBr(block)
			}
			c.SetInsertPoint(block)
			return parseStatementLLVM(l, c)
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

		// Just ensure the external is declared
		c.GetOrDeclareFunction(name)

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
func parseExpressionLLVM(l *Lexer, c *LLVMCompiler) (value.Value, error) {
	return parseAssignmentLLVM(l, c)
}

// parseAssignmentLLVM handles assignment expressions
func parseAssignmentLLVM(l *Lexer, c *LLVMCompiler) (value.Value, error) {
	// Try to parse the left side
	left, varName, err := parseAssignableLLVM(l, c)
	if err != nil {
		return nil, err
	}

	// Check for assignment operator
	if err := l.Whitespace(); err != nil {
		return nil, err
	}

	ch, err := l.ReadChar()
	if err == nil && ch == '=' {
		// This is an assignment
		ch2, err2 := l.ReadChar()
		if err2 == nil && ch2 == '=' {
			// It's == comparison, not assignment
			l.UnreadChar(ch2)
			l.UnreadChar(ch)
			return left, nil
		}
		if err2 == nil {
			l.UnreadChar(ch2)
		}

		// Parse the right side
		right, err := parseAssignmentLLVM(l, c)
		if err != nil {
			return nil, err
		}

		// Store the value
		if varName != "" {
			addr, err := c.GetAddress(varName)
			if err != nil {
				return nil, err
			}
			c.builder.NewStore(right, addr)
			return right, nil
		}
		return nil, fmt.Errorf("invalid assignment target")
	}

	if err == nil {
		l.UnreadChar(ch)
	}
	return left, nil
}

// parseAssignableLLVM parses an expression that can be assigned to
func parseAssignableLLVM(l *Lexer, c *LLVMCompiler) (value.Value, string, error) {
	// For now, just handle simple variable references
	if err := l.Whitespace(); err != nil {
		return nil, "", err
	}

	ch, err := l.ReadChar()
	if err != nil {
		return nil, "", err
	}

	if unicode.IsLetter(ch) {
		l.UnreadChar(ch)
		name, err := l.Identifier()
		if err != nil {
			return nil, "", err
		}

		// Check for function call
		if err := l.Whitespace(); err != nil {
			return nil, "", err
		}
		ch, err := l.ReadChar()
		if err == nil && ch == '(' {
			// Function call
			val, err := parseFunctionCallLLVM(l, c, name)
			return val, "", err
		}
		if err == nil {
			l.UnreadChar(ch)
		}

		// Variable - load it
		val, err := c.LoadValue(name)
		return val, name, err
	}

	// Not an identifier, try other expressions
	l.UnreadChar(ch)
	val, err := parseTermLLVM(l, c)
	return val, "", err
}

// parseTermLLVM parses a term (primary expression)
func parseTermLLVM(l *Lexer, c *LLVMCompiler) (value.Value, error) {
	if err := l.Whitespace(); err != nil {
		return nil, err
	}

	ch, err := l.ReadChar()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("unexpected end of file, expect expression")
		}
		return nil, err
	}

	switch {
	case ch == '\'':
		// Character literal
		val, err := l.Character()
		if err != nil {
			return nil, err
		}
		return constant.NewInt(c.WordType(), val), nil

	case ch == '"':
		// String literal
		str, err := l.String()
		if err != nil {
			return nil, err
		}
		global := c.CreateStringConstant(str)
		// Return pointer to first element
		gep := c.builder.NewGetElementPtr(global.ContentType, global,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0))
		return gep, nil

	case ch == '(':
		// Parentheses
		val, err := parseExpressionLLVM(l, c)
		if err != nil {
			return nil, err
		}
		if err := l.ExpectChar(')', "expect ')' after expression"); err != nil {
			return nil, err
		}
		return val, nil

	case unicode.IsDigit(ch):
		// Integer literal
		l.UnreadChar(ch)
		val, err := l.Number()
		if err != nil {
			return nil, err
		}
		return constant.NewInt(c.WordType(), val), nil

	case unicode.IsLetter(ch):
		// Identifier
		l.UnreadChar(ch)
		name, err := l.Identifier()
		if err != nil {
			return nil, err
		}

		// Check for function call
		if err := l.Whitespace(); err != nil {
			return nil, err
		}
		ch, err := l.ReadChar()
		if err == nil && ch == '(' {
			// Function call
			return parseFunctionCallLLVM(l, c, name)
		}
		if err == nil {
			l.UnreadChar(ch)
		}

		// Variable access
		return c.LoadValue(name)

	default:
		return nil, fmt.Errorf("unexpected character '%c', expect expression", ch)
	}
}

// parseFunctionCallLLVM parses a function call
func parseFunctionCallLLVM(l *Lexer, c *LLVMCompiler, name string) (value.Value, error) {
	fn := c.GetOrDeclareFunction(name)

	var args []value.Value
	for {
		ch, err := l.ReadChar()
		if err != nil {
			return nil, err
		}
		if ch == ')' {
			break
		}
		l.UnreadChar(ch)

		arg, err := parseExpressionLLVM(l, c)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if err := l.Whitespace(); err != nil {
			return nil, err
		}

		ch, err = l.ReadChar()
		if err != nil {
			return nil, err
		}
		if ch == ')' {
			break
		}
		if ch != ',' {
			return nil, fmt.Errorf("unexpected character '%c', expect ')' or ','", ch)
		}
	}

	call := c.builder.NewCall(fn, args...)
	return call, nil
}
