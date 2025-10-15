package main

import (
	"strings"
	"testing"
)

func TestLexerIdentifier(t *testing.T) {
	tests := []LexerTestConfig{
		{Name: "hello", Input: "hello", Want: "hello"},
		{Name: "main", Input: "main", Want: "main"},
		{Name: "test123", Input: "test123", Want: "test123"},
		{Name: "_var", Input: "_var", Want: "_var"},
		{Name: "var_name", Input: "var_name", Want: "var_name"},
		{Name: "x", Input: "x", Want: "x"},
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			runLexerTest(t, tt, func(l *Lexer) (interface{}, error) {
				return l.Identifier()
			})
		})
	}
}

func TestLexerNumber(t *testing.T) {
	tests := []LexerTestConfig{
		{Name: "42", Input: "42", Want: int64(42)},
		{Name: "0", Input: "0", Want: int64(0)},
		{Name: "123", Input: "123", Want: int64(123)},
		{Name: "010", Input: "010", Want: int64(8)},    // octal
		{Name: "077", Input: "077", Want: int64(63)},   // octal 77 = decimal 63
		{Name: "0123", Input: "0123", Want: int64(83)}, // octal 123 = decimal 83
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			runLexerTest(t, tt, func(l *Lexer) (interface{}, error) {
				return l.Number()
			})
		})
	}
}

func TestLexerString(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{"simple", `"hello"`, "hello"},
		{"empty", `""`, ""},
		{"with_spaces", `"hello world"`, "hello world"},
		{"with_escape_n", `"hello*nworld"`, "hello\nworld"},
		{"with_escape_t", `"tab*there"`, "tab\there"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			args := NewCompileOptions("test", nil)
			lexer := NewLexer(args, strings.NewReader(tt.Input))

			// Skip the opening quote
			lexer.ReadChar()

			got, err := lexer.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}
			if got != tt.Want {
				t.Errorf("String() = %q, want %q", got, tt.Want)
			}
		})
	}
}

func TestLexerCharacter(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  int64
	}{
		{"single", `'a'`, int64('a')},
		{"multi", `'ab'`, int64('a')<<8 | int64('b')},
		{"escape_n", `'*n'`, int64('\n')},
		{"escape_t", `'*t'`, int64('\t')},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			args := NewCompileOptions("test", nil)
			lexer := NewLexer(args, strings.NewReader(tt.Input))

			// Skip the opening quote
			lexer.ReadChar()

			got, err := lexer.Character()
			if err != nil {
				t.Fatalf("Character() error = %v", err)
			}
			if got != tt.Want {
				t.Errorf("Character() = %d, want %d", got, tt.Want)
			}
		})
	}
}

func TestLexerWhitespace(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  rune
	}{
		{"spaces", "   hello", 'h'},
		{"tabs", "\t\thello", 'h'},
		{"newlines", "\n\nhello", 'h'},
		{"mixed", " \t\n hello", 'h'},
		{"with_comment", "/* comment */ hello", 'h'},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			args := NewCompileOptions("test", nil)
			lexer := NewLexer(args, strings.NewReader(tt.Input))

			err := lexer.Whitespace()
			if err != nil {
				t.Fatalf("Whitespace() error = %v", err)
			}

			got, err := lexer.ReadChar()
			if err != nil {
				t.Fatalf("ReadChar() error = %v", err)
			}
			if got != tt.Want {
				t.Errorf("After Whitespace() got %c, want %c", got, tt.Want)
			}
		})
	}
}

func TestLexerComment(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{"simple", "/* comment */ rest", " rest"},
		{"multiline", "/* line1\nline2 */ rest", " rest"},
		{"nested_stars", "/* ** comment ** */ rest", " rest"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			args := NewCompileOptions("test", nil)
			lexer := NewLexer(args, strings.NewReader(tt.Input))

			// Skip the opening /*
			lexer.ReadChar()
			lexer.ReadChar()

			err := lexer.Comment()
			if err != nil {
				t.Fatalf("Comment() error = %v", err)
			}

			// Read the rest
			var rest []rune
			for {
				c, err := lexer.ReadChar()
				if err != nil {
					break
				}
				rest = append(rest, c)
			}

			got := string(rest)
			if got != tt.Want {
				t.Errorf("After Comment() got %q, want %q", got, tt.Want)
			}
		})
	}
}
