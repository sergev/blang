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

// parseExpressionWithLevel parses expressions with precedence level
func parseExpressionWithLevel(l *Lexer, c *Compiler, level int) (value.Value, error) {
	// Parse left side (term with unary operators)
	left, isLvalue, err := parseUnary(l, c)
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

		// Track if we handled an operator at this iteration
		handled := false

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
			thenVal, err := parseExpressionWithLevel(l, c, 12)
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
			elseVal, err := parseExpressionWithLevel(l, c, 13)
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
		if level >= 14 && ch == '=' && !handled {
			ch2, err2 := l.ReadChar()

			// Check for === first (compound assignment for equality)
			if err2 == nil && ch2 == '=' {
				ch3, err3 := l.ReadChar()
				if err3 == nil && ch3 == '=' {
					// === compound assignment
					if !isLvalue {
						return nil, fmt.Errorf("left operand of assignment must be an lvalue")
					}
					// Parse right side
					right, err := parseExpressionWithLevel(l, c, 14)
					if err != nil {
						return nil, err
					}
					// x === y → x = (x == y)
					currentVal := c.builder.NewLoad(c.WordType(), left)
					cmp := c.builder.NewICmp(enum.IPredEQ, currentVal, right)
					newVal := c.builder.NewZExt(cmp, c.WordType())
					c.builder.NewStore(newVal, left)
					return newVal, nil
				}

				// Not ===, check if it's == (equality comparison)
				if err3 == nil {
					l.UnreadChar(ch3)
				}
				// == equality comparison
				if isLvalue {
					left = c.builder.NewLoad(c.WordType(), left)
					isLvalue = false
				}
				if level >= 7 {
					right, err := parseExpressionWithLevel(l, c, 6)
					if err != nil {
						return nil, err
					}
					cmp := c.builder.NewICmp(enum.IPredEQ, left, right)
					left = c.builder.NewZExt(cmp, c.WordType())
					handled = true
					continue
				}
				l.UnreadChar(ch2)
				l.UnreadChar(ch)
				break
			}

			// Assignment (simple or compound)
			if !isLvalue {
				return nil, fmt.Errorf("left operand of assignment must be an lvalue")
			}

			// Check for compound assignment operators
			var compoundOp rune = 0
			if err2 == nil {
				// Check if it's a compound assignment
				switch ch2 {
				case '+', '-', '*', '/', '%', '&', '|':
					compoundOp = ch2
				case '<', '>':
					// Could be =<, =>, =<=, =>>, =>=
					ch3, err3 := l.ReadChar()
					if err3 == nil {
						if (ch2 == '<' && ch3 == '<') || (ch2 == '>' && ch3 == '>') {
							// =<< or =>>
							compoundOp = ch2
							// Mark as shift operator (use special marker)
							if ch2 == '<' {
								compoundOp = '«' // Left shift marker
							} else {
								compoundOp = '»' // Right shift marker
							}
						} else if ch3 == '=' {
							// =<= or =>=
							compoundOp = ch2
							// Mark as comparison with = (use special markers)
							if ch2 == '<' {
								compoundOp = '≤' // <= marker
							} else {
								compoundOp = '≥' // >= marker
							}
						} else {
							// Just =< or =>
							l.UnreadChar(ch3)
							compoundOp = ch2
						}
					} else {
						compoundOp = ch2
					}
				case '!':
					// =!=
					ch3, err3 := l.ReadChar()
					if err3 == nil && ch3 == '=' {
						compoundOp = '≠' // != marker
					} else {
						if err3 == nil {
							l.UnreadChar(ch3)
						}
						l.UnreadChar(ch2)
					}
				default:
					l.UnreadChar(ch2)
				}
			}

			// Parse right side
			right, err := parseExpressionWithLevel(l, c, 14)
			if err != nil {
				return nil, err
			}

			// Perform compound assignment or simple assignment
			if compoundOp != 0 {
				// Compound assignment: x =op y → x = x op y
				// Load current value
				currentVal := c.builder.NewLoad(c.WordType(), left)
				var newVal value.Value

				switch compoundOp {
				case '+':
					newVal = c.builder.NewAdd(currentVal, right)
				case '-':
					newVal = c.builder.NewSub(currentVal, right)
				case '*':
					newVal = c.builder.NewMul(currentVal, right)
				case '/':
					newVal = c.builder.NewSDiv(currentVal, right)
				case '%':
					newVal = c.builder.NewSRem(currentVal, right)
				case '&':
					newVal = c.builder.NewAnd(currentVal, right)
				case '|':
					newVal = c.builder.NewOr(currentVal, right)
				case '«': // =<<
					newVal = c.builder.NewShl(currentVal, right)
				case '»': // =>>
					newVal = c.builder.NewAShr(currentVal, right)
				case '<': // =<
					cmp := c.builder.NewICmp(enum.IPredSLT, currentVal, right)
					newVal = c.builder.NewZExt(cmp, c.WordType())
				case '≤': // =<=
					cmp := c.builder.NewICmp(enum.IPredSLE, currentVal, right)
					newVal = c.builder.NewZExt(cmp, c.WordType())
				case '>': // =>
					cmp := c.builder.NewICmp(enum.IPredSGT, currentVal, right)
					newVal = c.builder.NewZExt(cmp, c.WordType())
				case '≥': // =>=
					cmp := c.builder.NewICmp(enum.IPredSGE, currentVal, right)
					newVal = c.builder.NewZExt(cmp, c.WordType())
				case '≠': // =!=
					cmp := c.builder.NewICmp(enum.IPredNE, currentVal, right)
					newVal = c.builder.NewZExt(cmp, c.WordType())
				case '⩵': // ===
					cmp := c.builder.NewICmp(enum.IPredEQ, currentVal, right)
					newVal = c.builder.NewZExt(cmp, c.WordType())
				default:
					return nil, fmt.Errorf("unknown compound assignment operator")
				}
				c.builder.NewStore(newVal, left)
				return newVal, nil
			} else {
				// Simple assignment
				c.builder.NewStore(right, left)
				return right, nil
			}
		}

		// Binary operators (left associative)

		// Bitwise OR (level 10)
		if level >= 10 && ch == '|' && !handled {
			if isLvalue {
				left = c.builder.NewLoad(c.WordType(), left)
				isLvalue = false
			}
			right, err := parseExpressionWithLevel(l, c, 9)
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
			right, err := parseExpressionWithLevel(l, c, 7)
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
			right, err := parseExpressionWithLevel(l, c, 6)
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
						right, err = parseExpressionWithLevel(l, c, 4)
						if err != nil {
							return nil, err
						}
						left = c.builder.NewShl(left, right)
						handled = true
						continue
					}
				} else if err2 == nil && ch2 == '=' {
					// Less than or equal
					right, err = parseExpressionWithLevel(l, c, 5)
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
					right, err = parseExpressionWithLevel(l, c, 5)
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
						right, err = parseExpressionWithLevel(l, c, 4)
						if err != nil {
							return nil, err
						}
						left = c.builder.NewAShr(left, right)
						handled = true
						continue
					}
				} else if err2 == nil && ch2 == '=' {
					// Greater than or equal
					right, err = parseExpressionWithLevel(l, c, 5)
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
					right, err = parseExpressionWithLevel(l, c, 5)
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
			right, err := parseExpressionWithLevel(l, c, 3)
			if err != nil {
				return nil, err
			}

			// Note: In B, we can't distinguish pointers from integers at compile time
			// Pointer arithmetic scaling happens in the [] operator, not here
			// Regular addition/subtraction is just integer arithmetic
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
			right, err := parseExpressionWithLevel(l, c, 2)
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

// parseUnary parses unary operators and primary expressions
func parseUnary(l *Lexer, c *Compiler) (value.Value, bool, error) {
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
		val, isLvalue, err := parseUnary(l, c)
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
			val, isLvalue, err := parseUnary(l, c)
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
		val, isLvalue, err := parseUnary(l, c)
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
		val, isLvalue, err := parseUnary(l, c)
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
		val, isLvalue, err := parseUnary(l, c)
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
		val, isLvalue, err := parseUnary(l, c)
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
		return parsePostfix(l, c)
	}
}

// parsePostfix handles postfix operators and primary expressions
func parsePostfix(l *Lexer, c *Compiler) (value.Value, bool, error) {
	val, isLvalue, err := parsePrimary(l, c)
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
			// In B, array[i] means: (pointer + i * word_size)
			if isLvalue {
				val = c.builder.NewLoad(c.WordType(), val)
				isLvalue = false
			}

			// val is an i64 containing a pointer value
			// Convert to actual pointer type
			ptr := c.builder.NewIntToPtr(val, c.WordPtrType())

			// Parse index
			index, err := parseExpressionWithLevel(l, c, 15)
			if err != nil {
				return nil, false, err
			}
			if err := l.ExpectChar(']', "expect ']' after array index"); err != nil {
				return nil, false, err
			}

			// Calculate element address using getelementptr
			// This automatically scales by element size (i64 = 8 bytes)
			elemPtr := c.builder.NewGetElementPtr(c.WordType(), ptr, index)
			val = elemPtr
			isLvalue = true

		case '(':
			// Function call - handle both direct and indirect calls
			var fn value.Value

			if isLvalue {
				// It's a function pointer variable (from extrn declaration)
				// Load the function pointer value for indirect call
				fn = c.builder.NewLoad(c.WordType(), val)
				// Convert to function pointer type for indirect call
				// For now, we'll use inttoptr and call through a cast
			} else {
				// Direct function call
				fn = val
			}

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

				arg, err := parseExpressionWithLevel(l, c, 15)
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

			// Perform the call
			var result value.Value
			if fnDirect, ok := fn.(*ir.Func); ok {
				// Direct call to known function
				result = c.builder.NewCall(fnDirect, args...)
			} else {
				// Indirect call through function pointer
				// fn is the address of a variable containing the function address
				// Since it's marked as lvalue, load it to get the i64 function address
				fnAddr := c.builder.NewLoad(c.WordType(), fn)

				// Convert i64 to function pointer
				// Create variadic function type: i64 (i64, ...)*
				fnType := types.NewFunc(c.WordType())
				fnType.Variadic = true
				fnPtrType := types.NewPointer(fnType)
				fnPtr := c.builder.NewIntToPtr(fnAddr, fnPtrType)

				// Call through the pointer
				result = c.builder.NewCall(fnPtr, args...)
			}
			val = result
			isLvalue = false

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

// parsePrimary parses primary expressions (literals, identifiers, parentheses)
func parsePrimary(l *Lexer, c *Compiler) (value.Value, bool, error) {
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
		val, err := parseExpressionWithLevel(l, c, 15)
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

		// Peek ahead to see if this is a function call
		if err := l.Whitespace(); err != nil {
			return nil, false, err
		}
		nextCh, err := l.ReadChar()
		isCall := (err == nil && nextCh == '(')
		if err == nil {
			l.UnreadChar(nextCh)
		}

		// If it's a call, handle both direct and indirect calls
		if isCall {
			// Check module for a declared function
			if fn := c.findFuncByName(name); fn != nil {
				return fn, false, nil
			}

			// Check if it's an extrn variable (function pointer) in module globals
			if g := c.findGlobalByName(name); g != nil {
				return g, false, nil
			}

			// Not found anywhere - auto-declare as external function in current context
			fn := c.GetOrDeclareFunction(name)
			if fn == nil {
				return nil, false, fmt.Errorf("cannot declare function '%s'", name)
			}
			return fn, false, nil
		}

		// Not a call - get the variable address
		addr, found := c.GetAddress(name)
		if !found {
			return nil, false, fmt.Errorf("undefined identifier '%s'", name)
		}

		// Check if it's a function
		if fn, ok := addr.(*ir.Func); ok {
			// Function used as value (not called) - return its address as i64
			// This allows: func_ptr = add; (storing function address)
			fnPtr := c.builder.NewPtrToInt(fn, c.WordType())
			return fnPtr, false, nil
		}

		// It's a variable - return as lvalue
		return addr, true, nil

	default:
		return nil, false, fmt.Errorf("unexpected character '%c', expect expression", ch)
	}
}
