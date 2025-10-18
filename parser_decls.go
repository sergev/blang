package main

import (
	"fmt"
	"io"
	"unicode"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
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
			// Clear context after each top-level declaration
			c.ClearTopLevelContext()
		case '[':
			if err := parseVector(l, c, name); err != nil {
				return err
			}
			// Clear context after each top-level declaration
			c.ClearTopLevelContext()
		default:
			l.UnreadChar(ch)
			if err := parseGlobal(l, c, name); err != nil {
				return err
			}
			// Clear context after each top-level declaration
			c.ClearTopLevelContext()
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
	// Remove any existing global with the same name from the module
	c.removeGlobalByName(name)

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
	// Remove any existing global with the same name from the module
	c.removeGlobalByName(name)

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
		// TODO: Handle proper global references
		return nil, fmt.Errorf("initialization with global references is not supported yet")
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
