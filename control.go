package main

import (
	"fmt"
	"unicode"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
)

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
