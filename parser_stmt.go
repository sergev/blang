package main

import (
	"fmt"
	"unicode"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
)

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
			c.EnsureStackAlignment() // Align stack before first executable statement
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
		c.EnsureStackAlignment() // Align stack before first executable statement
		return parseReturn(l, c)
	case "auto":
		return parseAuto(l, c) // Declaration, not executable - no alignment needed
	case "extrn":
		return parseExtrn(l, c) // Declaration, not executable - no alignment needed
	case "if":
		c.EnsureStackAlignment() // Align stack before first executable statement
		return parseIf(l, c)
	case "while":
		c.EnsureStackAlignment() // Align stack before first executable statement
		return parseWhile(l, c)
	case "switch":
		c.EnsureStackAlignment() // Align stack before first executable statement
		return parseSwitch(l, c)
	case "case":
		c.EnsureStackAlignment() // Align stack before first executable statement
		return parseCase(l, c, switchID, cases)
	case "goto":
		c.EnsureStackAlignment() // Align stack before first executable statement
		return parseGoto(l, c)
	default:
		// Check if it's a label
		ch, err := l.ReadChar()
		if err != nil {
			return err
		}
		if ch == ':' {
			// Label - get or create labeled block
			c.EnsureStackAlignment() // Align stack before first executable statement
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
		c.EnsureStackAlignment() // Align stack before first executable statement
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

		// extrn declares a reference in the current declaration context only
		// Add a zero-initialized global to the module if not already present
		if c.findGlobalByName(name) == nil && c.findFuncByName(name) == nil {
			c.module.NewGlobalDef(c.globalName(name), constant.NewInt(c.WordType(), 0))
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

// parseGoto parses goto statements
func parseGoto(l *Lexer, c *Compiler) error {
	label, err := l.Identifier()
	if err != nil || label == "" {
		return fmt.Errorf("expect label name after 'goto'")
	}

	// Get or create the target label block
	targetBlock := c.GetOrCreateLabel(label)

	// Branch to the label
	c.builder.NewBr(targetBlock)

	// Create a new unreachable block for any code after the goto
	// This ensures subsequent code doesn't accidentally create branches
	deadBlock := c.NewBlock(fmt.Sprintf("unreachable.%d", c.labelID))
	c.labelID++
	c.SetInsertPoint(deadBlock)
	// Don't add unreachable instruction yet - let dead code be added,
	// but when we hit a label, we'll switch blocks

	if err := l.Whitespace(); err != nil {
		return err
	}
	return l.ExpectChar(';', "expect ';' after 'goto' statement")
}

// parseSwitch parses switch statements
func parseSwitch(l *Lexer, c *Compiler) error {
	switchID := c.labelID
	c.labelID++

	// Parse the switch expression
	switchVal, err := parseExpression(l, c)
	if err != nil {
		return err
	}

	// Create blocks
	stmtsBlock := c.NewBlock(fmt.Sprintf("switch.%d.stmts", switchID))
	cmpBlock := c.NewBlock(fmt.Sprintf("switch.%d.cmp", switchID))
	endBlock := c.NewBlock(fmt.Sprintf("switch.%d.end", switchID))

	// Jump to comparison block initially
	c.builder.NewBr(cmpBlock)

	// Parse the switch body (contains case statements)
	c.SetInsertPoint(stmtsBlock)
	var caseList []int64
	if err := parseStatementWithSwitch(l, c, int64(switchID), &caseList); err != nil {
		return err
	}

	// If no terminator, jump to end
	if c.builder.Term == nil {
		c.builder.NewBr(endBlock)
	}

	// Build the switch instruction in the comparison block
	c.SetInsertPoint(cmpBlock)
	if len(caseList) > 0 {
		// Create switch instruction
		sw := c.builder.NewSwitch(switchVal, endBlock)

		// Add all cases
		for _, caseVal := range caseList {
			caseBlock := c.GetOrCreateLabel(fmt.Sprintf("case.%d.%d", switchID, caseVal))
			sw.Cases = append(sw.Cases, ir.NewCase(constant.NewInt(c.WordType(), caseVal), caseBlock))
		}
	} else {
		// No cases, just jump to end
		c.builder.NewBr(endBlock)
	}

	// Continue with code after switch
	c.SetInsertPoint(endBlock)
	return nil
}

// parseCase parses case statements
func parseCase(l *Lexer, c *Compiler, switchID int64, cases *[]int64) error {
	if switchID < 0 {
		return fmt.Errorf("unexpected 'case' outside of 'switch' statements")
	}

	if cases == nil {
		return fmt.Errorf("invalid case list")
	}

	// Parse the case value
	var value int64
	ch, err := l.ReadChar()
	if err != nil {
		return err
	}

	switch ch {
	case '\'':
		value, err = l.Character()
		if err != nil {
			return err
		}
	default:
		if unicode.IsDigit(ch) {
			l.UnreadChar(ch)
			value, err = l.Number()
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unexpected character '%c', expect constant after 'case'", ch)
		}
	}

	if err := l.Whitespace(); err != nil {
		return err
	}
	if err := l.ExpectChar(':', "expect ':' after 'case'"); err != nil {
		return err
	}

	// Add to case list
	*cases = append(*cases, value)

	// Create label for this case
	caseBlock := c.GetOrCreateLabel(fmt.Sprintf("case.%d.%d", switchID, value))

	// Jump to case block if current block has no terminator
	if c.builder.Term == nil {
		c.builder.NewBr(caseBlock)
	}

	// Set insertion point to case block
	c.SetInsertPoint(caseBlock)

	// Parse the statement(s) following the case
	return parseStatementWithSwitch(l, c, switchID, cases)
}
