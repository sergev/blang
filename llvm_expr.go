package main

import (
	"fmt"
	"io"
	"unicode"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// parseExpressionLLVMWithLevel parses expressions with precedence level
func parseExpressionLLVMWithLevel(l *Lexer, c *LLVMCompiler, level int) (value.Value, error) {
	// Parse left side (term with unary operators)
	left, isLvalue, err := parseUnaryLLVM(l, c)
	if err != nil {
		return nil, err
	}

	// Dereference lvalue if needed for most operations
	for {
		if err := l.Whitespace(); err != nil {
			return nil, err
		}

		ch, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				if isLvalue {
					return c.builder.NewLoad(c.WordType(), left), nil
				}
				return left, nil
			}
			return nil, err
		}

		// Ternary operator (level 13)
		if level >= 13 && ch == '?' {
			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}

			condID := c.labelID
			c.labelID++
			thenBlock := c.NewBlock(fmt.Sprintf("cond.then.%d", condID))
			elseBlock := c.NewBlock(fmt.Sprintf("cond.else.%d", condID))
			endBlock := c.NewBlock(fmt.Sprintf("cond.end.%d", condID))

			// Branch on condition
			zero := constant.NewInt(c.WordType(), 0)
			cmp := c.builder.NewICmp(enum.IPredNE, left, zero)
			c.builder.NewCondBr(cmp, thenBlock, elseBlock)

			// Then branch
			c.SetInsertPoint(thenBlock)
			thenVal, err := parseExpressionLLVMWithLevel(l, c, 12)
			if err != nil {
				return nil, err
			}
			thenEndBlock := c.GetInsertBlock()
			c.builder.NewBr(endBlock)

			// Expect ':'
			if err := l.Whitespace(); err != nil {
				return nil, err
			}
			if err := l.ExpectChar(':', "expect ':' in ternary operator"); err != nil {
				return nil, err
			}

			// Else branch
			c.SetInsertPoint(elseBlock)
			elseVal, err := parseExpressionLLVMWithLevel(l, c, 13)
			if err != nil {
				return nil, err
			}
			elseEndBlock := c.GetInsertBlock()
			c.builder.NewBr(endBlock)

			// Merge
			c.SetInsertPoint(endBlock)
			phi := c.builder.NewPhi(ir.NewIncoming(thenVal, thenEndBlock), ir.NewIncoming(elseVal, elseEndBlock))
			return phi, nil
		}

		// Assignment operators (level 14, right associative)
		if level >= 14 && ch == '=' {
			ch2, err2 := l.ReadChar()
			if err2 == nil && ch2 == '=' {
				// Equality comparison ==, not assignment
				if isLvalue {
					left = c.builder.NewLoad(c.WordType(), left)
					isLvalue = false
				}
				if level >= 7 {
					right, err := parseExpressionLLVMWithLevel(l, c, 6)
					if err != nil {
						return nil, err
					}
					cmp := c.builder.NewICmp(enum.IPredEQ, left, right)
					return c.builder.NewZExt(cmp, c.WordType()), nil
				}
				l.UnreadChar(ch2)
				l.UnreadChar(ch)
				break
			}
			if err2 == nil {
				l.UnreadChar(ch2)
			}

			// Assignment
			if !isLvalue {
				return nil, fmt.Errorf("left operand of assignment must be an lvalue")
			}

			// TODO: Implement compound assignments (=+, =-, =*, etc.)
			// For now, only simple assignment (=) is supported
			// Use regular syntax: x = x + 5 instead of x =+ 5
			if err2 != nil {
				// Error reading second char, assume simple assignment
			}

			// Simple assignment
			right, err := parseExpressionLLVMWithLevel(l, c, 14)
			if err != nil {
				return nil, err
			}
			c.builder.NewStore(right, left)
			return right, nil
		}

		// Binary operators (left associative)
		handled := false

		// Bitwise OR (level 10)
		if level >= 10 && ch == '|' && !handled {
			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}
			right, err := parseExpressionLLVMWithLevel(l, c, 9)
			if err != nil {
				return nil, err
			}
			left = c.builder.NewOr(left, right)
			handled = true
			continue
		}

		// Bitwise AND (level 8)
		if level >= 8 && ch == '&' && !handled {
			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}
			right, err := parseExpressionLLVMWithLevel(l, c, 7)
			if err != nil {
				return nil, err
			}
			left = c.builder.NewAnd(left, right)
			handled = true
			continue
		}

		// Inequality (level 7)
		if level >= 7 && ch == '!' && !handled {
			ch2, err2 := l.ReadChar()
			if err2 != nil || ch2 != '=' {
				return nil, fmt.Errorf("unknown operator '!%c'", ch2)
			}
			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}
			right, err := parseExpressionLLVMWithLevel(l, c, 6)
			if err != nil {
				return nil, err
			}
			cmp := c.builder.NewICmp(enum.IPredNE, left, right)
			left = c.builder.NewZExt(cmp, c.WordType())
			handled = true
			continue
		}

		// Comparison operators (level 6)
		if level >= 6 && (ch == '<' || ch == '>') && !handled {
			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}

			ch2, err2 := l.ReadChar()
			var right value.Value

			if ch == '<' {
				if err2 == nil && ch2 == '<' {
					// Left shift (level 5)
					if level >= 5 {
						right, err = parseExpressionLLVMWithLevel(l, c, 4)
						if err != nil {
							return nil, err
						}
						left = c.builder.NewShl(left, right)
						handled = true
						continue
					}
				} else if err2 == nil && ch2 == '=' {
					// Less than or equal
					right, err = parseExpressionLLVMWithLevel(l, c, 5)
					if err != nil {
						return nil, err
					}
					cmp := c.builder.NewICmp(enum.IPredSLE, left, right)
					left = c.builder.NewZExt(cmp, c.WordType())
					handled = true
					continue
				} else {
					// Less than
					if err2 == nil {
						l.UnreadChar(ch2)
					}
					right, err = parseExpressionLLVMWithLevel(l, c, 5)
					if err != nil {
						return nil, err
					}
					cmp := c.builder.NewICmp(enum.IPredSLT, left, right)
					left = c.builder.NewZExt(cmp, c.WordType())
					handled = true
					continue
				}
			} else if ch == '>' {
				if err2 == nil && ch2 == '>' {
					// Right shift (level 5)
					if level >= 5 {
						right, err = parseExpressionLLVMWithLevel(l, c, 4)
						if err != nil {
							return nil, err
						}
						left = c.builder.NewAShr(left, right)
						handled = true
						continue
					}
				} else if err2 == nil && ch2 == '=' {
					// Greater than or equal
					right, err = parseExpressionLLVMWithLevel(l, c, 5)
					if err != nil {
						return nil, err
					}
					cmp := c.builder.NewICmp(enum.IPredSGE, left, right)
					left = c.builder.NewZExt(cmp, c.WordType())
					handled = true
					continue
				} else {
					// Greater than
					if err2 == nil {
						l.UnreadChar(ch2)
					}
					right, err = parseExpressionLLVMWithLevel(l, c, 5)
					if err != nil {
						return nil, err
					}
					cmp := c.builder.NewICmp(enum.IPredSGT, left, right)
					left = c.builder.NewZExt(cmp, c.WordType())
					handled = true
					continue
				}
			}
		}

		// Addition/Subtraction (level 4)
		if level >= 4 && (ch == '+' || ch == '-') && !handled {
			// Check for ++ or --
			ch2, _ := l.ReadChar()
			if ch2 == ch {
				// Postfix increment/decrement
				l.UnreadChar(ch2)
				l.UnreadChar(ch)
				break
			}
			if ch2 != 0 {
				l.UnreadChar(ch2)
			}

			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}
			right, err := parseExpressionLLVMWithLevel(l, c, 3)
			if err != nil {
				return nil, err
			}
			if ch == '+' {
				left = c.builder.NewAdd(left, right)
			} else {
				left = c.builder.NewSub(left, right)
			}
			handled = true
			continue
		}

		// Multiplication/Division/Modulo (level 3)
		if level >= 3 && (ch == '*' || ch == '/' || ch == '%') && !handled {
			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}
			right, err := parseExpressionLLVMWithLevel(l, c, 2)
			if err != nil {
				return nil, err
			}
			switch ch {
			case '*':
				left = c.builder.NewMul(left, right)
			case '/':
				left = c.builder.NewSDiv(left, right)
			case '%':
				left = c.builder.NewSRem(left, right)
			}
			handled = true
			continue
		}

		// No operator found at this level
		if !handled {
			l.UnreadChar(ch)
			break
		}
	}

	// Dereference lvalue if needed
	if isLvalue {
		return c.builder.NewLoad(c.WordType(), left), nil
	}
	return left, nil
}

// parseUnaryLLVM parses unary operators and primary expressions
func parseUnaryLLVM(l *Lexer, c *LLVMCompiler) (value.Value, bool, error) {
	if err := l.Whitespace(); err != nil {
		return nil, false, err
	}

	ch, err := l.ReadChar()
	if err != nil {
		if err == io.EOF {
			return nil, false, fmt.Errorf("unexpected end of file, expect expression")
		}
		return nil, false, err
	}

	switch ch {
	case '!':
		// Logical NOT
		val, isLvalue, err := parseUnaryLLVM(l, c)
		if err != nil {
			return nil, false, err
		}
		if isLvalue {
			val = c.builder.NewLoad(c.WordType(), val)
		}
		zero := constant.NewInt(c.WordType(), 0)
		cmp := c.builder.NewICmp(enum.IPredEQ, val, zero)
		// Convert i1 to i64
		result := c.builder.NewZExt(cmp, c.WordType())
		return result, false, nil

	case '-':
		// Check for prefix decrement --
		ch2, err2 := l.ReadChar()
		if err2 == nil && ch2 == '-' {
			// Prefix decrement
			val, isLvalue, err := parseUnaryLLVM(l, c)
			if err != nil {
				return nil, false, err
			}
			if !isLvalue {
				return nil, false, fmt.Errorf("expected lvalue after '--'")
			}
			current := c.builder.NewLoad(c.WordType(), val)
			one := constant.NewInt(c.WordType(), 1)
			result := c.builder.NewSub(current, one)
			c.builder.NewStore(result, val)
			return val, true, nil
		}
		if err2 == nil {
			l.UnreadChar(ch2)
		}

		// Negation
		val, isLvalue, err := parseUnaryLLVM(l, c)
		if err != nil {
			return nil, false, err
		}
		if isLvalue {
			val = c.builder.NewLoad(c.WordType(), val)
		}
		zero := constant.NewInt(c.WordType(), 0)
		result := c.builder.NewSub(zero, val)
		return result, false, nil

	case '+':
		// Prefix increment
		ch2, err2 := l.ReadChar()
		if err2 != nil || ch2 != '+' {
			return nil, false, fmt.Errorf("unexpected character '%c', expect '+'", ch2)
		}
		val, isLvalue, err := parseUnaryLLVM(l, c)
		if err != nil {
			return nil, false, err
		}
		if !isLvalue {
			return nil, false, fmt.Errorf("expected lvalue after '++'")
		}
		current := c.builder.NewLoad(c.WordType(), val)
		one := constant.NewInt(c.WordType(), 1)
		result := c.builder.NewAdd(current, one)
		c.builder.NewStore(result, val)
		return val, true, nil

	case '*':
		// Indirection (dereference)
		val, isLvalue, err := parseUnaryLLVM(l, c)
		if err != nil {
			return nil, false, err
		}
		if isLvalue {
			val = c.builder.NewLoad(c.WordType(), val)
		}
		// val now contains a pointer (i64), treat as lvalue
		// Cast to pointer type
		ptr := c.builder.NewIntToPtr(val, c.WordPtrType())
		return ptr, true, nil

	case '&':
		// Address-of
		val, isLvalue, err := parseUnaryLLVM(l, c)
		if err != nil {
			return nil, false, err
		}
		if !isLvalue {
			return nil, false, fmt.Errorf("expected lvalue after '&'")
		}
		// Convert pointer to integer
		result := c.builder.NewPtrToInt(val, c.WordType())
		return result, false, nil

	default:
		l.UnreadChar(ch)
		return parsePostfixLLVM(l, c)
	}
}

// parsePostfixLLVM handles postfix operators and primary expressions
func parsePostfixLLVM(l *Lexer, c *LLVMCompiler) (value.Value, bool, error) {
	val, isLvalue, err := parsePrimaryLLVM(l, c)
	if err != nil {
		return nil, false, err
	}

	for {
		if err := l.Whitespace(); err != nil {
			return nil, false, err
		}

		ch, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return val, isLvalue, nil
			}
			return nil, false, err
		}

		switch ch {
		case '[':
			// Array indexing
			if isLvalue {
				val = c.builder.NewLoad(c.WordType(), val)
				isLvalue = false
			}
			// Load pointer from array variable
			ptr := c.builder.NewIntToPtr(val, c.WordPtrType())

			// Parse index
			index, err := parseExpressionLLVMWithLevel(l, c, 15)
			if err != nil {
				return nil, false, err
			}
			if err := l.ExpectChar(']', "expect ']' after array index"); err != nil {
				return nil, false, err
			}

			// Calculate address: ptr + index
			elemPtr := c.builder.NewGetElementPtr(c.WordType(), ptr, index)
			val = elemPtr
			isLvalue = true

		case '(':
			// Function call
			if isLvalue {
				val = c.builder.NewLoad(c.WordType(), val)
				isLvalue = false
			}

			// val should be a function or function pointer
			// For now, we handle it as getting the function by name
			// This is a simplification - proper implementation would handle function pointers

			var args []value.Value
			for {
				ch, err := l.ReadChar()
				if err != nil {
					return nil, false, err
				}
				if ch == ')' {
					break
				}
				l.UnreadChar(ch)

				arg, err := parseExpressionLLVMWithLevel(l, c, 15)
				if err != nil {
					return nil, false, err
				}
				args = append(args, arg)

				if err := l.Whitespace(); err != nil {
					return nil, false, err
				}
				ch, err = l.ReadChar()
				if err != nil {
					return nil, false, err
				}
				if ch == ')' {
					break
				}
				if ch != ',' {
					return nil, false, fmt.Errorf("unexpected character '%c', expect ')'", ch)
				}
			}

			// Try to convert val to function
			// This is a workaround - proper implementation needed
			if fn, ok := val.(*ir.Func); ok {
				result := c.builder.NewCall(fn, args...)
				val = result
				isLvalue = false
			} else {
				return nil, false, fmt.Errorf("cannot call non-function value")
			}

		case '+':
			// Postfix increment
			ch2, err2 := l.ReadChar()
			if err2 != nil || ch2 != '+' {
				if err2 == nil {
					l.UnreadChar(ch2)
				}
				l.UnreadChar(ch)
				return val, isLvalue, nil
			}
			if !isLvalue {
				return nil, false, fmt.Errorf("expected lvalue for postfix '++'")
			}
			current := c.builder.NewLoad(c.WordType(), val)
			one := constant.NewInt(c.WordType(), 1)
			result := c.builder.NewAdd(current, one)
			c.builder.NewStore(result, val)
			val = current // Return old value
			isLvalue = false

		case '-':
			// Postfix decrement
			ch2, err2 := l.ReadChar()
			if err2 != nil || ch2 != '-' {
				if err2 == nil {
					l.UnreadChar(ch2)
				}
				l.UnreadChar(ch)
				return val, isLvalue, nil
			}
			if !isLvalue {
				return nil, false, fmt.Errorf("expected lvalue for postfix '--'")
			}
			current := c.builder.NewLoad(c.WordType(), val)
			one := constant.NewInt(c.WordType(), 1)
			result := c.builder.NewSub(current, one)
			c.builder.NewStore(result, val)
			val = current // Return old value
			isLvalue = false

		default:
			l.UnreadChar(ch)
			return val, isLvalue, nil
		}
	}
}

// parsePrimaryLLVM parses primary expressions (literals, identifiers, parentheses)
func parsePrimaryLLVM(l *Lexer, c *LLVMCompiler) (value.Value, bool, error) {
	if err := l.Whitespace(); err != nil {
		return nil, false, err
	}

	ch, err := l.ReadChar()
	if err != nil {
		if err == io.EOF {
			return nil, false, fmt.Errorf("unexpected end of file, expect expression")
		}
		return nil, false, err
	}

	switch {
	case ch == '\'':
		// Character literal
		val, err := l.Character()
		if err != nil {
			return nil, false, err
		}
		return constant.NewInt(c.WordType(), val), false, nil

	case ch == '"':
		// String literal
		str, err := l.String()
		if err != nil {
			return nil, false, err
		}
		global := c.CreateStringConstant(str)
		gep := c.builder.NewGetElementPtr(global.ContentType, global,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0))
		// Cast to i64
		result := c.builder.NewPtrToInt(gep, c.WordType())
		return result, false, nil

	case ch == '(':
		// Parenthesized expression
		val, err := parseExpressionLLVMWithLevel(l, c, 15)
		if err != nil {
			return nil, false, err
		}
		if err := l.ExpectChar(')', "expect ')' after expression"); err != nil {
			return nil, false, err
		}
		return val, false, nil

	case unicode.IsDigit(ch):
		// Integer literal
		l.UnreadChar(ch)
		val, err := l.Number()
		if err != nil {
			return nil, false, err
		}
		return constant.NewInt(c.WordType(), val), false, nil

	case unicode.IsLetter(ch):
		// Identifier
		l.UnreadChar(ch)
		name, err := l.Identifier()
		if err != nil {
			return nil, false, err
		}

		// Try to get address first (for lvalue)
		addr, found := c.GetAddress(name)
		if !found {
			// Check if it's followed by '(' - if so, it's an external function
			if err := l.Whitespace(); err != nil {
				return nil, false, err
			}
			ch, err := l.ReadChar()
			if err == nil {
				if ch == '(' {
					// Declare as external function
					l.UnreadChar(ch)
					fn := c.GetOrDeclareFunction(name)
					return fn, false, nil
				}
				l.UnreadChar(ch)
			}
			// Not a function call - it's an undefined variable
			return nil, false, fmt.Errorf("undefined identifier '%s'", name)
		}

		// Check if it's a function
		if fn, ok := addr.(*ir.Func); ok {
			return fn, false, nil
		}

		// It's a variable - return as lvalue
		return addr, true, nil

	default:
		return nil, false, fmt.Errorf("unexpected character '%c', expect expression", ch)
	}
}
