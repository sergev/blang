package main

import (
	"fmt"
	"io"
	"unicode"
)

// Lexer handles tokenization and input reading
type Lexer struct {
	args   *CompileOptions
	reader io.RuneReader
	buffer []rune // pushback buffer for unread characters
}

// NewLexer creates a new lexer
func NewLexer(args *CompileOptions, reader io.Reader) *Lexer {
	return &Lexer{
		args:   args,
		reader: &runeReaderAdapter{reader},
		buffer: make([]rune, 0),
	}
}

// runeReaderAdapter adapts io.Reader to io.RuneReader
type runeReaderAdapter struct {
	reader io.Reader
}

func (r *runeReaderAdapter) ReadRune() (rune, int, error) {
	buf := make([]byte, 1)
	n, err := r.reader.Read(buf)
	if err != nil {
		return 0, 0, err
	}
	if n == 0 {
		return 0, 0, io.EOF
	}
	return rune(buf[0]), 1, nil
}

// ReadChar reads one character from input
func (l *Lexer) ReadChar() (rune, error) {
	// Check pushback buffer first
	if len(l.buffer) > 0 {
		c := l.buffer[len(l.buffer)-1]
		l.buffer = l.buffer[:len(l.buffer)-1]
		return c, nil
	}

	c, _, err := l.reader.ReadRune()
	return c, err
}

// UnreadChar pushes a character back to be read again
func (l *Lexer) UnreadChar(c rune) {
	l.buffer = append(l.buffer, c)
}

// Comment parses a comment (/* ... */)
func (l *Lexer) Comment() error {
	for {
		c, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("unclosed comment, expect '*/' to close the comment")
			}
			return err
		}

		if c == '*' {
			c2, err := l.ReadChar()
			if err != nil && err != io.EOF {
				return err
			}
			if c2 == '/' {
				return nil
			}
			if err != io.EOF {
				l.UnreadChar(c2)
			}
		}
	}
}

// Whitespace skips whitespace and comments
func (l *Lexer) Whitespace() error {
	for {
		c, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if unicode.IsSpace(c) {
			continue
		}

		if c == '/' {
			c2, err := l.ReadChar()
			if err != nil {
				if err == io.EOF {
					l.UnreadChar(c)
					return nil
				}
				return err
			}

			if c2 == '*' {
				if err := l.Comment(); err != nil {
					return err
				}
				continue
			} else {
				l.UnreadChar(c2)
				l.UnreadChar(c)
				return nil
			}
		}

		l.UnreadChar(c)
		return nil
	}
}

// Identifier parses an identifier (alphanumeric + underscore)
func (l *Lexer) Identifier() (string, error) {
	if err := l.Whitespace(); err != nil {
		return "", err
	}

	var result []rune

	for {
		c, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return string(result), nil
			}
			return "", err
		}

		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			l.UnreadChar(c)
			return string(result), nil
		}

		result = append(result, c)
	}
}

// Number parses an integer literal (octal if starts with 0)
func (l *Lexer) Number() (int64, error) {
	if err := l.Whitespace(); err != nil {
		return 0, err
	}

	c, err := l.ReadChar()
	if err != nil {
		return 0, err
	}

	base := 10
	if c == '0' {
		base = 8
	}

	var num int64 = 0
	for unicode.IsDigit(c) {
		num = (num * int64(base)) + int64(c-'0')
		c, err = l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return num, nil
			}
			return 0, err
		}
	}

	l.UnreadChar(c)
	return num, nil
}

// Escape parses an escape character
func (l *Lexer) Escape() (rune, error) {
	c, err := l.ReadChar()
	if err != nil {
		return 0, err
	}

	switch c {
	case '0', 'e':
		return '\000', nil
	case '*', '\'', '"':
		return c, nil
	case '(':
		return '{', nil
	case ')':
		return '}', nil
	case 't':
		return '\t', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	default:
		return 0, fmt.Errorf("undefined escape character '*%c'", c)
	}
}

// Character parses a multi-character literal
func (l *Lexer) Character() (int64, error) {
	var value int64 = 0

	for i := 0; i < l.args.WordSize; i++ {
		c, err := l.ReadChar()
		if err != nil {
			return 0, err
		}

		if c == '\'' {
			return value, nil
		}

		if c == '*' {
			c, err = l.Escape()
			if err != nil {
				return 0, err
			}
		}

		// Big endian
		value = (value << 8) | int64(c&0xFF)
	}

	c, err := l.ReadChar()
	if err != nil || c != '\'' {
		return 0, fmt.Errorf("unclosed char literal")
	}

	return value, nil
}

// String parses a string literal
func (l *Lexer) String() (string, error) {
	var result []rune

	for {
		c, err := l.ReadChar()
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("unterminated string literal")
			}
			return "", err
		}

		if c == '"' {
			return string(result), nil
		}

		if c == '*' {
			c, err = l.Escape()
			if err != nil {
				return "", err
			}
		}

		result = append(result, c)
	}
}

// PeekChar looks at the next character without consuming it
func (l *Lexer) PeekChar() (rune, error) {
	c, err := l.ReadChar()
	if err != nil {
		return 0, err
	}
	l.UnreadChar(c)
	return c, nil
}

// ExpectChar reads a character and checks if it matches expected
func (l *Lexer) ExpectChar(expected rune, msg string) error {
	c, err := l.ReadChar()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("%s", msg)
		}
		return err
	}

	if c != expected {
		return fmt.Errorf("%s, got '%c'", msg, c)
	}

	return nil
}

// IsEOF checks if we're at end of file
func (l *Lexer) IsEOF() bool {
	c, err := l.ReadChar()
	if err == io.EOF {
		return true
	}
	if err == nil {
		l.UnreadChar(c)
	}
	return false
}
