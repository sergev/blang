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

// ParseDeclarations parses top-level declarations and generates LLVM IR
func ParseDeclarations(l *Lexer, c *Compiler) error {
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
			if err := parseFunction(l, c, name); err != nil {
				return err
			}
		case '[':
			if err := parseVector(l, c, name); err != nil {
				return err
			}
		default:
			l.UnreadChar(ch)
			if err := parseGlobal(l, c, name); err != nil {
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

// parseGlobal parses a global variable
func parseGlobal(l *Lexer, c *Compiler, name string) error {
	// If this name was already declared as extrn (as external reference),
	// remove it so we can declare it properly
	if existing, exists := c.globals[name]; exists {
		// Remove from globals map
		delete(c.globals, name)
		// Remove from module's globals list
		for i, g := range c.module.Globals {
			if g == existing {
				c.module.Globals = append(c.module.Globals[:i], c.module.Globals[i+1:]...)
				break
			}
		}
	}

	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	if ch != ';' {
		l.UnreadChar(ch)
		// Parse initialization list
		var initVals []constant.Constant
		for {
			if err := l.Whitespace(); err != nil {
				return err
			}
			val, err := parseIvalConst(l, c)
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

		// If multiple values, allocate multiple words for scalar
		// (not an array - just a scalar with consecutive initialization)
		if len(initVals) > 1 {
			c.DeclareGlobalWithMultipleValues(name, initVals)
		} else {
			c.DeclareGlobal(name, initVals[0])
		}
	} else {
		c.DeclareGlobal(name, nil)
	}

	return nil
}

// parseVector parses a global array
func parseVector(l *Lexer, c *Compiler, name string) error {
	// If this name was already declared as extrn (as a scalar global),
	// remove it so we can declare it as an array
	if existing, exists := c.globals[name]; exists {
		// Remove from globals map
		delete(c.globals, name)
		// Remove from module's globals list
		for i, g := range c.module.Globals {
			if g == existing {
				c.module.Globals = append(c.module.Globals[:i], c.module.Globals[i+1:]...)
				break
			}
		}
	}

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
			val, err := parseIvalConst(l, c)
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

// parseIvalConst parses a constant initialization value
func parseIvalConst(l *Lexer, c *Compiler) (constant.Constant, error) {
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
		// Get pointer to first element of string constant using GEP
		gep := constant.NewGetElementPtr(global.ContentType, global,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0))
		// Convert string pointer to i64 for array storage
		return constant.NewPtrToInt(gep, c.WordType()), nil
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

// parseFunction parses a function definition
func parseFunction(l *Lexer, c *Compiler, name string) error {
	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	var paramNames []string
	if ch != ')' {
		l.UnreadChar(ch)
		paramNames, err = parseArguments(l)
		if err != nil {
			return err
		}
	}

	fn := c.DeclareFunction(name, paramNames)
	c.StartFunction(fn)

	if err := parseStatementWithSwitch(l, c, -1, nil); err != nil {
		return err
	}

	c.EndFunction()
	return nil
}

// parseArguments parses function arguments
func parseArguments(l *Lexer) ([]string, error) {
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

// parseStatement parses a statement
// switchID: ID of enclosing switch statement (-1 if none)
// cases: slice to accumulate case values for switch
func parseStatement(l *Lexer, c *Compiler) error {
	return parseStatementWithSwitch(l, c, -1, nil)
}

func parseStatementWithSwitch(l *Lexer, c *Compiler, switchID int64, cases *[]int64) error {
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
			if err := parseStatementWithSwitch(l, c, switchID, cases); err != nil {
				return err
			}
		}

	case ';':
		// Null statement

	default:
		if unicode.IsLetter(ch) {
			l.UnreadChar(ch)
			return parseKeywordOrExpressionWithSwitch(l, c, switchID, cases)
		} else {
			l.UnreadChar(ch)
			// Expression statement
			_, err := parseExpression(l, c)
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

func parseKeywordOrExpressionWithSwitch(l *Lexer, c *Compiler, switchID int64, cases *[]int64) error {
	name, err := l.Identifier()
	if err != nil {
		return err
	}

	if err := l.Whitespace(); err != nil {
		return err
	}

	switch name {
	case "return":
		return parseReturn(l, c)
	case "auto":
		return parseAuto(l, c)
	case "extrn":
		return parseExtrn(l, c)
	case "if":
		return parseIf(l, c)
	case "while":
		return parseWhile(l, c)
	case "switch":
		return parseSwitch(l, c)
	case "case":
		return parseCase(l, c, switchID, cases)
	case "goto":
		return parseGoto(l, c)
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
			return parseStatementWithSwitch(l, c, switchID, cases)
		}

		// Otherwise it's an expression
		l.UnreadChar(ch)
		for i := len(name) - 1; i >= 0; i-- {
			l.UnreadChar(rune(name[i]))
		}
		_, err = parseExpression(l, c)
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

// parseReturn parses a return statement
func parseReturn(l *Lexer, c *Compiler) error {
	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	if ch != ';' {
		if ch != '(' {
			return fmt.Errorf("expect '(' or ';' after 'return'")
		}
		val, err := parseExpression(l, c)
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

// parseAuto parses auto variable declarations
// Note: Variables within ONE auto statement are allocated in forward order.
// Reverse allocation happens at the statement level (last statement allocated first).
func parseAuto(l *Lexer, c *Compiler) error {
	// Collect all declarations in this auto statement
	type autoDecl struct {
		name string
		size int64 // -1 for scalar, >= 0 for array
	}
	var decls []autoDecl

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
			if err := l.Whitespace(); err != nil {
				return err
			}

			ch2, err := l.ReadChar()
			if err != nil {
				return err
			}

			var size int64 = 0
			if ch2 != ']' {
				// Array size can be a number or character constant
				if ch2 == '\'' {
					// Character constant (already read the opening ')
					size, err = l.Character()
					if err != nil {
						return err
					}
				} else {
					// Should be a number
					l.UnreadChar(ch2)
					size, err = l.Number()
					if err != nil {
						return err
					}
				}
				if err := l.Whitespace(); err != nil {
					return err
				}
				if err := l.ExpectChar(']', "expect ']' after array size"); err != nil {
					return err
				}
			}

			decls = append(decls, autoDecl{name: name, size: size})

			if err := l.Whitespace(); err != nil {
				return err
			}
			ch, err = l.ReadChar()
			if err != nil {
				return err
			}

			if ch == ';' {
				break
			}
			if ch != ',' {
				return fmt.Errorf("unexpected character '%c', expect ';' or ','", ch)
			}
		} else {
			// Scalar variable - no initialization allowed
			// Must be ';' or ','
			if ch != ';' && ch != ',' {
				return fmt.Errorf("unexpected character '%c', expect ';' or ',' after auto variable", ch)
			}
			decls = append(decls, autoDecl{name: name, size: -1})
			if ch == ';' {
				break
			}
		}
	}

	// Declare in FORWARD order (variables within one auto statement)
	// The reverse order happens at the statement level, not variable level
	for i := 0; i < len(decls); i++ {
		decl := decls[i]
		if decl.size == -1 {
			// Scalar
			c.DeclareLocal(decl.name)
		} else {
			// Array
			c.DeclareLocalArray(decl.name, decl.size)
		}
	}

	return nil
}

// parseExtrn parses external declarations
func parseExtrn(l *Lexer, c *Compiler) error {
	for {
		name, err := l.Identifier()
		if err != nil || name == "" {
			return fmt.Errorf("expect identifier after 'extrn'")
		}

		// In B, extrn declares a reference to an external name (variable or function)
		// We create a forward reference as an external global (uninitialized)
		// If a definition appears later in the file, it will replace this
		//
		// Examples:
		//   extrn printf;  printf("hello");  → printf will be auto-declared as function
		//   extrn n; ... n 2000;             → n defined later will replace this
		//   extrn v; ... v[2000];            → v defined later will replace this

		if _, exists := c.globals[name]; !exists {
			if _, exists := c.functions[name]; !exists {
				// Declare as external global variable (i64) - may be replaced later
				// Initialize to 0 for proper LLVM IR
				global := c.module.NewGlobalDef(name, constant.NewInt(c.WordType(), 0))
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

// parseIf parses if statements
func parseIf(l *Lexer, c *Compiler) error {
	if err := l.ExpectChar('(', "expect '(' after 'if'"); err != nil {
		return err
	}

	cond, err := parseExpression(l, c)
	if err != nil {
		return err
	}

	if err := l.Whitespace(); err != nil {
		return err
	}
	if err := l.ExpectChar(')', "expect ')' after condition"); err != nil {
		return err
	}

	// Create blocks with unique IDs to avoid label conflicts in nested if-else
	ifID := c.labelID
	c.labelID++
	thenBlock := c.NewBlock(fmt.Sprintf("if.%d.then", ifID))
	elseBlock := c.NewBlock(fmt.Sprintf("if.%d.else", ifID))
	endBlock := c.NewBlock(fmt.Sprintf("if.%d.end", ifID))

	// Compare condition to zero
	zero := constant.NewInt(c.WordType(), 0)
	cmp := c.builder.NewICmp(enum.IPredNE, cond, zero)
	c.builder.NewCondBr(cmp, thenBlock, elseBlock)

	// Generate then block
	c.SetInsertPoint(thenBlock)
	if err := parseStatement(l, c); err != nil {
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
		if err := parseStatement(l, c); err != nil {
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

// parseWhile parses while loops
func parseWhile(l *Lexer, c *Compiler) error {
	if err := l.ExpectChar('(', "expect '(' after 'while'"); err != nil {
		return err
	}

	// Create unique blocks for this while loop
	whileID := c.labelID
	c.labelID++
	condBlock := c.NewBlock(fmt.Sprintf("while.%d.cond", whileID))
	bodyBlock := c.NewBlock(fmt.Sprintf("while.%d.body", whileID))
	endBlock := c.NewBlock(fmt.Sprintf("while.%d.end", whileID))

	// Jump to condition
	c.builder.NewBr(condBlock)
	c.SetInsertPoint(condBlock)

	// Evaluate condition
	cond, err := parseExpression(l, c)
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
	if err := parseStatement(l, c); err != nil {
		return err
	}
	if c.builder.Term == nil {
		c.builder.NewBr(condBlock)
	}

	c.SetInsertPoint(endBlock)
	return nil
}

// parseExpression parses an expression and returns the result value
// This is a wrapper that calls the comprehensive expression parser with full precedence support
func parseExpression(l *Lexer, c *Compiler) (value.Value, error) {
	return parseExpressionWithLevel(l, c, 15)
}
