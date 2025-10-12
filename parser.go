package main

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

// errorf creates a formatted error
func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// ParseDeclarations parses top-level declarations
func ParseDeclarations(l *Lexer) ([]byte, error) {
	var out bytes.Buffer

	for {
		name, err := l.Identifier()
		if err != nil {
			return nil, err
		}
		if name == "" {
			break
		}

		fmt.Fprintf(&out, ".globl %s\n", name)

		c, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return nil, errorf("unexpected end of file after declaration")
			}
			return nil, err
		}

		switch c {
		case '(':
			if err := parseFunction(l, &out, name); err != nil {
				return nil, err
			}
		case '[':
			if err := parseVector(l, &out, name); err != nil {
				return nil, err
			}
		default:
			l.UnreadChar(c)
			if err := parseGlobal(l, &out, name); err != nil {
				return nil, err
			}
		}
	}

	// Check for unexpected input
	_, err := l.ReadChar()
	if err != io.EOF {
		if err == nil {
			return nil, errorf("expect identifier at top level")
		}
		return nil, err
	}

	// Add string section
	writeStrings(l.args, &out)

	return out.Bytes(), nil
}

// parseGlobal parses a global scalar variable declaration
func parseGlobal(l *Lexer, out *bytes.Buffer, name string) error {
	fmt.Fprintf(out,
		".data\n"+
			".type %s, @object\n"+
			".align %d\n"+
			"%s:\n",
		name, l.args.WordSize, name,
	)

	c, err := l.ReadChar()
	if err != nil {
		return err
	}

	if c != ';' {
		l.UnreadChar(c)
		for {
			if err := l.Whitespace(); err != nil {
				return err
			}
			if err := parseIval(l, out); err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}

			c, err := l.ReadChar()
			if err != nil {
				return err
			}
			if c == ';' {
				break
			}
			if c != ',' {
				return errorf("expect ';' at end of declaration")
			}
		}
	} else {
		fmt.Fprintf(out, "  .zero %d\n", l.args.WordSize)
	}

	return nil
}

// parseVector parses a global array (vector) declaration
func parseVector(l *Lexer, out *bytes.Buffer, name string) error {
	var nwords int64 = 0

	if err := l.Whitespace(); err != nil {
		return err
	}

	c, err := l.ReadChar()
	if err != nil {
		return err
	}

	if c != ']' {
		l.UnreadChar(c)
		nwords, err = l.Number()
		if err != nil {
			return errorf("unexpected end of file, expect vector size after '['")
		}

		if err := l.Whitespace(); err != nil {
			return err
		}

		if err := l.ExpectChar(']', "expect ']' after vector size"); err != nil {
			return err
		}
	}

	fmt.Fprintf(out,
		".data\n.type %s, @object\n"+
			".align %d\n"+
			"%s:\n"+
			"  .quad .+8\n",
		name, l.args.WordSize, name,
	)

	if err := l.Whitespace(); err != nil {
		return err
	}

	c, err = l.ReadChar()
	if err != nil {
		return err
	}

	if c != ';' {
		l.UnreadChar(c)
		for {
			if err := l.Whitespace(); err != nil {
				return err
			}
			if err := parseIval(l, out); err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			nwords--

			c, err := l.ReadChar()
			if err != nil {
				return err
			}
			if c == ';' {
				break
			}
			if c != ',' {
				return errorf("expect ';' at end of declaration")
			}
		}
	}

	if nwords > 0 {
		fmt.Fprintf(out, "  .zero %d\n", l.args.WordSize*int(nwords))
	}

	return nil
}

// parseIval parses an initialization value
func parseIval(l *Lexer, out *bytes.Buffer) error {
	c, err := l.ReadChar()
	if err != nil {
		return err
	}

	if unicode.IsLetter(c) {
		l.UnreadChar(c)
		name, err := l.Identifier()
		if err != nil || name == "" {
			return errorf("unexpected end of file, expect ival")
		}
		fmt.Fprintf(out, "  .quad %s\n", name)
	} else if c == '\'' {
		value, err := l.Character()
		if err != nil {
			return errorf("unexpected end of file, expect ival")
		}
		fmt.Fprintf(out, "  .quad %d\n", value)
	} else if c == '"' {
		str, err := l.String()
		if err != nil {
			return err
		}
		l.args.Strings.Push(str)
		fmt.Fprintf(out, "  .quad .string.%d\n", l.args.Strings.Size-1)
	} else if c == '-' {
		value, err := l.Number()
		if err != nil {
			return errorf("unexpected end of file, expect ival")
		}
		fmt.Fprintf(out, "  .quad -%d\n", value)
	} else {
		l.UnreadChar(c)
		value, err := l.Number()
		if err != nil {
			return errorf("unexpected end of file, expect ival")
		}
		fmt.Fprintf(out, "  .quad %d\n", value)
	}

	return nil
}

// parseFunction parses a function definition
func parseFunction(l *Lexer, out *bytes.Buffer, fnIdent string) error {
	// Clear locals
	l.args.Locals.Clear()
	l.args.StackOffset = 0

	// Clear externs
	l.args.Extrns.Clear()

	// Add function name to externs
	l.args.Extrns.Push(fnIdent)

	fmt.Fprintf(out,
		".text\n"+
			".type %s, @function\n"+
			"%s:\n"+
			"  push %%rbp\n"+
			"  mov %%rsp, %%rbp\n"+
			"  sub $%d, %%rsp\n",
		fnIdent, fnIdent, l.args.WordSize,
	)

	c, err := l.ReadChar()
	if err != nil {
		return err
	}

	if c != ')' {
		l.UnreadChar(c)
		if err := parseArguments(l, out); err != nil {
			return err
		}
	}

	if err := parseStatement(l, out, fnIdent, -1, nil); err != nil {
		return err
	}

	fmt.Fprintf(out,
		"  xor %%rax, %%rax\n"+
			".L.return.%s:\n"+
			"  mov %%rbp, %%rsp\n"+
			"  pop %%rbp\n"+
			"  ret\n",
		fnIdent,
	)

	return nil
}

// parseArguments parses function arguments
func parseArguments(l *Lexer, out *bytes.Buffer) error {
	argIndex := 0

	for {
		if err := l.Whitespace(); err != nil {
			return err
		}

		name, err := l.Identifier()
		if err != nil || name == "" {
			return errorf("expect ')' or identifier after function arguments")
		}

		fmt.Fprintf(out, "  sub $%d, %%rsp\n  mov %s, -%d(%%rbp)\n",
			l.args.WordSize, ArgRegisters[argIndex], (l.args.StackOffset+2)*uint64(l.args.WordSize))
		argIndex++

		l.args.Locals.Push(&StackVar{Name: name, Offset: l.args.StackOffset})
		l.args.StackOffset++

		if err := l.Whitespace(); err != nil {
			return err
		}

		c, err := l.ReadChar()
		if err != nil {
			return err
		}

		switch c {
		case ')':
			return nil
		case ',':
			continue
		default:
			return errorf("unexpected character '%c', expect ')' or ','\n", c)
		}
	}
}

// parseStatement parses a statement
func parseStatement(l *Lexer, out *bytes.Buffer, fnIdent string, switchID int64, cases *List) error {
	if err := l.Whitespace(); err != nil {
		return err
	}

	c, err := l.ReadChar()
	if err != nil {
		return err
	}

	switch c {
	case '{':
		stackOffset := l.args.StackOffset

		for {
			if err := l.Whitespace(); err != nil {
				return err
			}
			c, err := l.ReadChar()
			if err != nil {
				return err
			}
			if c == '}' {
				break
			}
			l.UnreadChar(c)
			if err := parseStatement(l, out, fnIdent, switchID, cases); err != nil {
				return err
			}
		}

		// Reset stack
		if stackOffset != l.args.StackOffset {
			fmt.Fprintf(out, "  add $%d, %%rsp\n", (l.args.StackOffset-stackOffset)*uint64(l.args.WordSize))
			l.args.StackOffset = stackOffset
		}

	case ';':
		// Null statement

	default:
		if unicode.IsLetter(c) {
			l.UnreadChar(c)
			return parseKeywordOrExpression(l, out, fnIdent, switchID, cases)
		} else {
			l.UnreadChar(c)
			if err := parseExpression(l, out, 15); err != nil {
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

// parseKeywordOrExpression handles keywords and expressions starting with identifiers
func parseKeywordOrExpression(l *Lexer, out *bytes.Buffer, fnIdent string, switchID int64, cases *List) error {
	name, err := l.Identifier()
	if err != nil {
		return err
	}

	if err := l.Whitespace(); err != nil {
		return err
	}

	switch name {
	case "goto":
		return parseGoto(l, out, fnIdent)
	case "return":
		return parseReturn(l, out, fnIdent)
	case "if":
		return parseIf(l, out, fnIdent)
	case "while":
		return parseWhile(l, out, fnIdent)
	case "switch":
		return parseSwitch(l, out, fnIdent)
	case "case":
		return parseCase(l, out, fnIdent, switchID, cases)
	case "extrn":
		return parseExtrn(l, out)
	case "auto":
		return parseAuto(l, out)
	default:
		// Check if it's a label
		c, err := l.ReadChar()
		if err != nil {
			return err
		}
		if c == ':' {
			fmt.Fprintf(out, ".L.label.%s.%s:\n", name, fnIdent)
			return parseStatement(l, out, fnIdent, switchID, cases)
		}

		// Otherwise it's an expression
		l.UnreadChar(c)
		for i := len(name) - 1; i >= 0; i-- {
			l.UnreadChar(rune(name[i]))
		}
		if err := parseExpression(l, out, 15); err != nil {
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

// Keyword statement parsers

func parseGoto(l *Lexer, out *bytes.Buffer, fnIdent string) error {
	label, err := l.Identifier()
	if err != nil || label == "" {
		return errorf("expect label name after 'goto'")
	}
	fmt.Fprintf(out, "  jmp .L.label.%s.%s\n", label, fnIdent)
	if err := l.Whitespace(); err != nil {
		return err
	}
	return l.ExpectChar(';', "expect ';' after 'goto' statement")
}

func parseReturn(l *Lexer, out *bytes.Buffer, fnIdent string) error {
	c, err := l.ReadChar()
	if err != nil {
		return err
	}

	if c != ';' {
		if c != '(' {
			return errorf("expect '(' or ';' after 'return'")
		}
		if err := parseExpression(l, out, 15); err != nil {
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
	} else {
		fmt.Fprintf(out, "  xor %%rax, %%rax\n")
	}

	fmt.Fprintf(out, "  jmp .L.return.%s\n", fnIdent)
	return nil
}

func parseIf(l *Lexer, out *bytes.Buffer, fnIdent string) error {
	id := l.args.StmtCnt
	l.args.StmtCnt++

	if err := l.ExpectChar('(', "expect '(' after 'if'"); err != nil {
		return err
	}
	if err := parseExpression(l, out, 15); err != nil {
		return err
	}
	fmt.Fprintf(out, "  cmp $0, %%rax\n  je .L.else.%d\n", id)

	if err := l.Whitespace(); err != nil {
		return err
	}
	if err := l.ExpectChar(')', "expect ')' after condition"); err != nil {
		return err
	}

	if err := parseStatement(l, out, fnIdent, -1, nil); err != nil {
		return err
	}

	fmt.Fprintf(out, "  jmp .L.end.%d\n.L.else.%d:\n", id, id)

	// Check for else
	if err := l.Whitespace(); err != nil {
		return err
	}

	// Try to read "else"
	elseChars := []rune{'e', 'l', 's', 'e'}
	var readChars []rune
	isElse := true

	for _, expected := range elseChars {
		c, err := l.ReadChar()
		if err != nil || c != expected {
			isElse = false
			if err == nil {
				readChars = append(readChars, c)
			}
			break
		}
		readChars = append(readChars, c)
	}

	if isElse {
		// Check that next char is not alphanumeric
		c, err := l.ReadChar()
		if err == nil {
			readChars = append(readChars, c)
			if unicode.IsLetter(c) || unicode.IsDigit(c) {
				isElse = false
			}
		}
	}

	if isElse {
		if err := parseStatement(l, out, fnIdent, -1, nil); err != nil {
			return err
		}
	} else {
		// Push back characters
		for i := len(readChars) - 1; i >= 0; i-- {
			l.UnreadChar(readChars[i])
		}
	}

	fmt.Fprintf(out, ".L.end.%d:\n", id)
	return nil
}

func parseWhile(l *Lexer, out *bytes.Buffer, fnIdent string) error {
	id := l.args.StmtCnt
	l.args.StmtCnt++

	if err := l.ExpectChar('(', "expect '(' after 'while'"); err != nil {
		return err
	}
	fmt.Fprintf(out, ".L.start.%d:\n", id)
	if err := parseExpression(l, out, 15); err != nil {
		return err
	}
	fmt.Fprintf(out, "  cmp $0, %%rax\n  je .L.end.%d\n", id)

	if err := l.Whitespace(); err != nil {
		return err
	}
	if err := l.ExpectChar(')', "expect ')' after condition"); err != nil {
		return err
	}

	if err := parseStatement(l, out, fnIdent, -1, nil); err != nil {
		return err
	}

	fmt.Fprintf(out, "  jmp .L.start.%d\n.L.end.%d:\n", id, id)
	return nil
}

func parseSwitch(l *Lexer, out *bytes.Buffer, fnIdent string) error {
	id := l.args.StmtCnt
	l.args.StmtCnt++

	if err := parseExpression(l, out, 15); err != nil {
		return err
	}
	fmt.Fprintf(out, "  jmp .L.cmp.%d\n.L.stmts.%d:\n", id, id)

	switchCaseList := NewList()
	if err := parseStatement(l, out, fnIdent, int64(id), switchCaseList); err != nil {
		return err
	}

	fmt.Fprintf(out, "  jmp .L.end.%d\n.L.cmp.%d:\n", id, id)

	for i := 0; i < switchCaseList.Size; i++ {
		caseVal := switchCaseList.Data[i].(int64)
		fmt.Fprintf(out, "  cmp $%d, %%rax\n  je .L.case.%d.%d\n", caseVal, id, caseVal)
	}

	fmt.Fprintf(out, ".L.end.%d:\n", id)
	return nil
}

func parseCase(l *Lexer, out *bytes.Buffer, fnIdent string, switchID int64, cases *List) error {
	if switchID < 0 {
		return errorf("unexpected 'case' outside of 'switch' statements")
	}

	var value int64
	c, err := l.ReadChar()
	if err != nil {
		return err
	}

	switch c {
	case '\'':
		value, err = l.Character()
		if err != nil {
			return err
		}
	default:
		if unicode.IsDigit(c) {
			l.UnreadChar(c)
			value, err = l.Number()
			if err != nil {
				return err
			}
		} else {
			return errorf("unexpected character '%c', expect constant after 'case'\n", c)
		}
	}

	if err := l.Whitespace(); err != nil {
		return err
	}
	if err := l.ExpectChar(':', "expect ':' after 'case'"); err != nil {
		return err
	}

	cases.Push(value)
	fmt.Fprintf(out, ".L.case.%d.%d:\n", switchID, value)
	return parseStatement(l, out, fnIdent, switchID, cases)
}

func parseExtrn(l *Lexer, out *bytes.Buffer) error {
	for {
		name, err := l.Identifier()
		if err != nil || name == "" {
			return errorf("expect identifier after 'extrn'")
		}

		_, _, found := l.args.FindIdentifier(name)
		if found {
			return errorf("identifier '%s' is already defined in this scope\n", name)
		}

		l.args.Extrns.Push(name)

		if err := l.Whitespace(); err != nil {
			return err
		}

		c, err := l.ReadChar()
		if err != nil {
			return err
		}

		if c == ';' {
			return nil
		}
		if c != ',' {
			return errorf("unexpected character '%c', expect ';' or ','\n", c)
		}
	}
}

func parseAuto(l *Lexer, out *bytes.Buffer) error {
	for {
		name, err := l.Identifier()
		if err != nil || name == "" {
			return errorf("expect identifier after 'auto'")
		}

		_, _, found := l.args.FindIdentifier(name)
		if found {
			return errorf("identifier '%s' is already defined in this scope\n", name)
		}

		if err := l.Whitespace(); err != nil {
			return err
		}

		value := int64(-1)
		c, err := l.ReadChar()
		if err != nil {
			return err
		}

		if c == '\'' {
			value, err = l.Character()
			if err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			c, err = l.ReadChar()
			if err != nil {
				return err
			}
		} else if c == '[' {
			value, err = l.Number()
			if err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			if err := l.ExpectChar(']', "unexpected character, expect ']'"); err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			c, err = l.ReadChar()
			if err != nil {
				return err
			}
		} else if unicode.IsDigit(c) {
			l.UnreadChar(c)
			value, err = l.Number()
			if err != nil {
				return err
			}
			if err := l.Whitespace(); err != nil {
				return err
			}
			c, err = l.ReadChar()
			if err != nil {
				return err
			}
		}

		if value < 0 {
			// Scalar
			l.args.Locals.Push(&StackVar{Name: name, Offset: l.args.StackOffset})
			l.args.StackOffset++
			fmt.Fprintf(out, "  sub $%d, %%rsp\n", l.args.WordSize)
		} else {
			// Vector
			l.args.Locals.Push(&StackVar{Name: name, Offset: l.args.StackOffset + uint64(value)})
			l.args.StackOffset += uint64(value) + 1
			fmt.Fprintf(out, "  sub $%d, %%rsp\n", l.args.WordSize*int(value+1))

			// Initialize pointer
			fmt.Fprintf(out, "  lea -%d(%%rbp), %%rax\n", l.args.StackOffset*uint64(l.args.WordSize))
			fmt.Fprintf(out, "  movq %%rax, -%d(%%rbp)\n", (l.args.StackOffset+1)*uint64(l.args.WordSize))
		}

		if c == ';' {
			break
		}
		if c != ',' {
			return errorf("unexpected character '%c', expect ';' or ','\n", c)
		}
	}

	// Align stack to 16 bytes
	if l.args.StackOffset%2 != 0 {
		fmt.Fprintf(out, "  sub $%d, %%rsp\n", l.args.WordSize)
		l.args.StackOffset++
	}

	return nil
}

// writeStrings creates read-only section with strings
func writeStrings(args *CompilerArgs, out *bytes.Buffer) {
	if args.Strings.Size == 0 {
		return
	}

	fmt.Fprintf(out, ".section .rodata\n")

	for i := 0; i < args.Strings.Size; i++ {
		fmt.Fprintf(out, ".string.%d:\n", i)
		str := args.Strings.Data[i].(string)
		for _, c := range str {
			fmt.Fprintf(out, "  .byte %d\n", c)
		}
		fmt.Fprintf(out, "  .byte 0\n")
	}
}
