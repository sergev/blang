package main

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

// parseExpression parses an expression with the given precedence level
func parseExpression(l *Lexer, out *bytes.Buffer, level int) error {
	leftIsLvalue, err := parseTerm(l, out)
	if err != nil {
		return err
	}

	for {
		if err := l.Whitespace(); err != nil {
			return err
		}

		c, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				if leftIsLvalue {
					fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				}
				return nil
			}
			return err
		}

		// Ternary operator (lowest precedence)
		if level >= 13 && c == '?' {
			condID := l.args.ConditionalCnt
			l.args.ConditionalCnt++

			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			fmt.Fprintf(out, "  cmp $0, %%rax\n  je .L.cond.else.%d\n", condID)
			if err := parseExpression(l, out, 12); err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			c2, err := l.ReadChar()
			if err != nil || c2 != ':' {
				return errorf("unexpected character, expect ':' between conditional branches")
			}
			fmt.Fprintf(out, "  jmp .L.cond.end.%d\n.L.cond.else.%d:\n", condID, condID)
			if err := parseExpression(l, out, 13); err != nil {
				return err
			}
			fmt.Fprintf(out, ".L.cond.end.%d:\n", condID)
			return nil
		}

		// Binary operators
		handled := false

		// Addition (precedence 4)
		if level >= 4 && c == '+' {
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := binaryExpr(l, out, BinAdd, 3); err != nil {
				return err
			}
			handled = true
		}

		// Subtraction (precedence 4)
		if level >= 4 && c == '-' && !handled {
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := binaryExpr(l, out, BinSub, 3); err != nil {
				return err
			}
			handled = true
		}

		// Multiplication (precedence 3)
		if level >= 3 && c == '*' && !handled {
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := binaryExpr(l, out, BinMul, 2); err != nil {
				return err
			}
			handled = true
		}

		// Division (precedence 3)
		if level >= 3 && c == '/' && !handled {
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := binaryExpr(l, out, BinDiv, 2); err != nil {
				return err
			}
			handled = true
		}

		// Modulo (precedence 3)
		if level >= 3 && c == '%' && !handled {
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := binaryExpr(l, out, BinMod, 2); err != nil {
				return err
			}
			handled = true
		}

		// Shift and comparison operators starting with '<'
		if c == '<' && !handled {
			c2, err := l.ReadChar()
			if err != nil && err != io.EOF {
				return err
			}

			if level >= 5 && c2 == '<' {
				// Shift-left
				if leftIsLvalue {
					fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
					leftIsLvalue = false
				}
				if err := binaryExpr(l, out, BinShl, 4); err != nil {
					return err
				}
				handled = true
			} else if level >= 6 && c2 == '=' {
				// Less-than-equal
				if leftIsLvalue {
					fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
					leftIsLvalue = false
				}
				if err := cmpExpr(l, out, CmpLE, 5); err != nil {
					return err
				}
				handled = true
			} else if level >= 6 {
				// Less-than
				if err == nil {
					l.UnreadChar(c2)
				}
				if leftIsLvalue {
					fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
					leftIsLvalue = false
				}
				if err := cmpExpr(l, out, CmpLT, 5); err != nil {
					return err
				}
				handled = true
			} else {
				if err == nil {
					l.UnreadChar(c2)
				}
			}
		}

		// Shift and comparison operators starting with '>'
		if c == '>' && !handled {
			c2, err := l.ReadChar()
			if err != nil && err != io.EOF {
				return err
			}

			if level >= 5 && c2 == '>' {
				// Shift-right
				if leftIsLvalue {
					fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
					leftIsLvalue = false
				}
				if err := binaryExpr(l, out, BinSar, 4); err != nil {
					return err
				}
				handled = true
			} else if level >= 6 && c2 == '=' {
				// Greater-than-equal
				if leftIsLvalue {
					fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
					leftIsLvalue = false
				}
				if err := cmpExpr(l, out, CmpGE, 5); err != nil {
					return err
				}
				handled = true
			} else if level >= 6 {
				// Greater-than
				if err == nil {
					l.UnreadChar(c2)
				}
				if leftIsLvalue {
					fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
					leftIsLvalue = false
				}
				if err := cmpExpr(l, out, CmpGT, 5); err != nil {
					return err
				}
				handled = true
			} else {
				if err == nil {
					l.UnreadChar(c2)
				}
			}
		}

		// Inequality operator
		if level >= 7 && c == '!' && !handled {
			c2, err := l.ReadChar()
			if err != nil || c2 != '=' {
				return errorf("unknown operator '!%c'\n", c2)
			}
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := cmpExpr(l, out, CmpNE, 6); err != nil {
				return err
			}
			handled = true
		}

		// Bitwise AND
		if level >= 8 && c == '&' && !handled {
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := binaryExpr(l, out, BinAnd, 7); err != nil {
				return err
			}
			handled = true
		}

		// Bitwise OR
		if level >= 10 && c == '|' && !handled {
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
				leftIsLvalue = false
			}
			if err := binaryExpr(l, out, BinOr, 9); err != nil {
				return err
			}
			handled = true
		}

		// Assignment operators
		if c == '=' && !handled {
			c2, err := l.ReadChar()
			if err != nil && err != io.EOF {
				return err
			}

			if level >= 7 && c2 == '=' {
				// Check for === (third =)
				c3, err3 := l.ReadChar()
				if err3 == nil {
					l.UnreadChar(c3)
				}
				if err3 != nil || c3 != '=' {
					// Equality operator ==
					if leftIsLvalue {
						fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
						leftIsLvalue = false
					}
					if err := cmpExpr(l, out, CmpEQ, 6); err != nil {
						return err
					}
					handled = true
				}
			}

			if level >= 14 && !handled {
				// Assignment operator
				if !leftIsLvalue {
					return errorf("left operand of assignment has to be an lvalue")
				}
				fmt.Fprintf(out, "  push %%rax\n  mov (%%rax), %%rax\n")
				if err := assignExpr(l, out, c2, 14); err != nil {
					return err
				}
				fmt.Fprintf(out, "  pop %%rdi\n  mov %%rax, (%%rdi)\n")
				leftIsLvalue = false
				handled = true
			} else if !handled {
				if err == nil {
					l.UnreadChar(c2)
				}
			}
		}

		if !handled {
			// No more operations at this level
			l.UnreadChar(c)
			if leftIsLvalue {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
			}
			return nil
		}
	}
}

// parseTerm parses a term (primary expression with unary operators)
func parseTerm(l *Lexer, out *bytes.Buffer) (bool, error) {
	if err := l.Whitespace(); err != nil {
		return false, err
	}

	c, err := l.ReadChar()
	if err != nil {
		if err == io.EOF {
			return false, errorf("unexpected end of file, expect expression")
		}
		return false, err
	}

	isLvalue := false

	switch {
	case c == '\'':
		// Character literal
		value, err := l.Character()
		if err != nil {
			return false, err
		}
		if value != 0 {
			fmt.Fprintf(out, "  mov $%d, %%rax\n", value)
		} else {
			fmt.Fprintf(out, "  xor %%rax, %%rax\n")
		}

	case c == '"':
		// String literal
		str, err := l.String()
		if err != nil {
			return false, err
		}
		l.args.Strings.Push(str)
		fmt.Fprintf(out, "  lea .string.%d(%%rip), %%rax\n", l.args.Strings.Size-1)

	case c == '(':
		// Parentheses
		if err := parseExpression(l, out, 15); err != nil {
			return false, err
		}
		if err := l.ExpectChar(')', "expect ')' after '(<expr>', got '%c'"); err != nil {
			return false, err
		}

	case c == '!':
		// Not operator
		lval, err := parseTerm(l, out)
		if err != nil {
			return false, err
		}
		if lval {
			fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
		}
		fmt.Fprintf(out, "  cmp $0, %%rax\n  sete %%al\n  movzx %%al, %%rax\n")

	case c == '-':
		// Negation or prefix decrement
		c2, err := l.ReadChar()
		if err != nil && err != io.EOF {
			return false, err
		}

		if c2 == '-' {
			// Prefix decrement
			lval, err := parseTerm(l, out)
			if err != nil {
				return false, err
			}
			if !lval {
				return false, errorf("expected lvalue after '--'")
			}
			fmt.Fprintf(out, "  mov (%%rax), %%rdi\n  sub $1, %%rdi\n  mov %%rdi, (%%rax)\n")
			isLvalue = true
		} else {
			// Negation
			if err == nil {
				l.UnreadChar(c2)
			}
			lval, err := parseTerm(l, out)
			if err != nil {
				return false, err
			}
			if lval {
				fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
			}
			fmt.Fprintf(out, "  neg %%rax\n")
		}

	case c == '+':
		// Prefix increment
		c2, err := l.ReadChar()
		if err != nil || c2 != '+' {
			return false, errorf("unexpected character '%c', expect '+'\n", c2)
		}
		lval, err := parseTerm(l, out)
		if err != nil {
			return false, err
		}
		if !lval {
			return false, errorf("expected lvalue after '++'")
		}
		fmt.Fprintf(out, "  mov (%%rax), %%rdi\n  add $1, %%rdi\n  mov %%rdi, (%%rax)\n")
		isLvalue = true

	case c == '*':
		// Indirection operator
		lval, err := parseTerm(l, out)
		if err != nil {
			return false, err
		}
		if lval {
			fmt.Fprintf(out, "  mov (%%rax), %%rax\n")
		}
		isLvalue = true

	case c == '&':
		// Address operator
		lval, err := parseTerm(l, out)
		if err != nil {
			return false, err
		}
		if !lval {
			return false, errorf("expected lvalue after '&'")
		}

	case unicode.IsDigit(c):
		// Integer literal
		l.UnreadChar(c)
		value, err := l.Number()
		if err != nil {
			return false, err
		}
		if value != 0 {
			fmt.Fprintf(out, "  mov $%d, %%rax\n", value)
		} else {
			fmt.Fprintf(out, "  xor %%rax, %%rax\n")
		}

	case unicode.IsLetter(c):
		// Identifier
		l.UnreadChar(c)
		name, err := l.Identifier()
		if err != nil {
			return false, err
		}

		offset, isExtrn, found := l.args.FindIdentifier(name)

		if !found {
			// Unknown identifier - check if it's a function call
			if err := l.Whitespace(); err != nil {
				return false, err
			}
			c, err := l.ReadChar()
			if err == nil && c == '(' {
				// Add to externs
				l.UnreadChar(c)
				l.args.Extrns.Push(name)
				isExtrn = true
			} else {
				return false, errorf("undefined identifier '%s'\n", name)
			}
		}

		if isExtrn {
			fmt.Fprintf(out, "  lea %s(%%rip), %%rax\n", name)
		} else {
			fmt.Fprintf(out, "  lea -%d(%%rbp), %%rax\n", (offset+2)*int64(l.args.WordSize))
		}

		isLvalue = true
		isLvalue, err = parsePostfix(l, out, isLvalue)
		if err != nil {
			return false, err
		}

	default:
		return false, errorf("unexpected character '%c', expect expression\n", c)
	}

	return isLvalue, nil
}

// parsePostfix handles postfix operations ([], (), ++, --)
func parsePostfix(l *Lexer, out *bytes.Buffer, isLvalue bool) (bool, error) {
	for {
		c, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return isLvalue, nil
			}
			return false, err
		}

		switch c {
		case '[':
			// Index operator
			fmt.Fprintf(out, "  push (%%rax)\n")
			if err := parseExpression(l, out, 15); err != nil {
				return false, err
			}
			fmt.Fprintf(out, "  pop %%rdi\n  shl $3, %%rax\n  add %%rdi, %%rax\n")

			if err := l.ExpectChar(']', "unexpected token, expect closing ']' after index expression"); err != nil {
				return false, err
			}
			isLvalue = true

		case '(':
			// Function call
			fmt.Fprintf(out, "  push %%rax\n")

			numArgs := 0
			for {
				c, err := l.ReadChar()
				if err != nil {
					return false, err
				}
				if c == ')' {
					break
				}
				l.UnreadChar(c)

				if err := parseExpression(l, out, 15); err != nil {
					return false, err
				}

				numArgs++
				if numArgs > MaxFnCallArgs {
					return false, errorf("only %d call arguments are currently supported\n", MaxFnCallArgs)
				}
				fmt.Fprintf(out, "  push %%rax\n")

				if err := l.Whitespace(); err != nil {
					return false, err
				}

				c, err = l.ReadChar()
				if err != nil {
					return false, err
				}
				if c == ')' {
					break
				}
				if c != ',' {
					return false, errorf("unexpected character '%c', expect closing ')' after call expression\n", c)
				}
			}

			// Pop arguments into registers
			for numArgs > 0 {
				numArgs--
				fmt.Fprintf(out, "  pop %s\n", ArgRegisters[numArgs])
			}

			fmt.Fprintf(out, "  pop %%r10\n  call *%%r10\n")
			isLvalue = false

		case '+':
			// Postfix increment
			c2, err := l.ReadChar()
			if err != nil || c2 != '+' {
				if err == nil {
					l.UnreadChar(c2)
				}
				l.UnreadChar(c)
				return isLvalue, nil
			}
			fmt.Fprintf(out, "  mov (%%rax), %%rcx\n  addq $1, (%%rax)\n  mov %%rcx, %%rax\n")
			isLvalue = false

		case '-':
			// Postfix decrement
			c2, err := l.ReadChar()
			if err != nil || c2 != '-' {
				if err == nil {
					l.UnreadChar(c2)
				}
				l.UnreadChar(c)
				return isLvalue, nil
			}
			fmt.Fprintf(out, "  mov (%%rax), %%rcx\n  subq $1, (%%rax)\n  mov %%rcx, %%rax\n")
			isLvalue = false

		default:
			l.UnreadChar(c)
			return isLvalue, nil
		}
	}
}

// binaryExpr generates code for a binary operation
func binaryExpr(l *Lexer, out *bytes.Buffer, op BinaryOperator, level int) error {
	fmt.Fprintf(out, "  push %%rax\n")
	if err := parseExpression(l, out, level); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", BinaryCode[op])
	return nil
}

// cmpExpr generates code for a comparison operation
func cmpExpr(l *Lexer, out *bytes.Buffer, op CmpOperator, level int) error {
	fmt.Fprintf(out, "  push %%rax\n")
	if err := parseExpression(l, out, level); err != nil {
		return err
	}
	fmt.Fprintf(out,
		"  pop %%rdi\n"+
			"  cmp %%rax, %%rdi\n"+
			"  %s %%al\n"+
			"  movzb %%al, %%rax\n",
		CmpInstruction[op],
	)
	return nil
}

// assignExpr handles assignment operations (=, =+, =-, etc.)
func assignExpr(l *Lexer, out *bytes.Buffer, c rune, level int) error {
	switch c {
	case '+':
		return binaryExpr(l, out, BinAdd, level)
	case '*':
		return binaryExpr(l, out, BinMul, level)
	case '-':
		return binaryExpr(l, out, BinSub, level)
	case '/':
		return binaryExpr(l, out, BinDiv, level)
	case '%':
		return binaryExpr(l, out, BinMod, level)

	case '<':
		c2, err := l.ReadChar()
		if err != nil {
			return err
		}
		if c2 == '<' {
			// Shift-left
			return binaryExpr(l, out, BinShl, level)
		} else if c2 == '=' {
			// Less-than-equal
			return cmpExpr(l, out, CmpLE, level)
		} else {
			// Less-than
			l.UnreadChar(c2)
			return cmpExpr(l, out, CmpLT, level)
		}

	case '>':
		c2, err := l.ReadChar()
		if err != nil {
			return err
		}
		if c2 == '>' {
			// Shift-right
			return binaryExpr(l, out, BinSar, level)
		} else if c2 == '=' {
			// Greater-than-equal
			return cmpExpr(l, out, CmpGE, level)
		} else {
			// Greater-than
			l.UnreadChar(c2)
			return cmpExpr(l, out, CmpGT, level)
		}

	case '!':
		c2, err := l.ReadChar()
		if err != nil || c2 != '=' {
			return errorf("unknown operator '!%c'\n", c2)
		}
		return cmpExpr(l, out, CmpNE, level)

	case '=':
		c2, err := l.ReadChar()
		if err != nil || c2 != '=' {
			return errorf("unknown operator '=%c'\n", c2)
		}
		return cmpExpr(l, out, CmpEQ, level)

	case '&':
		return binaryExpr(l, out, BinAnd, level)

	case '|':
		return binaryExpr(l, out, BinOr, level)

	default:
		// Plain assignment
		l.UnreadChar(c)
		return parseExpression(l, out, level)
	}
}
