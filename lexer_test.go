package main

import (
	"testing"
)

func TestLexerIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"main", "main"},
		{"test123", "test123"},
		{"_var", "_var"},
		{"var_name", "var_name"},
		{"x", "x"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := lexIdentifier(t, tt.input)
			if got != tt.want {
				t.Errorf("Identifier() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexerNumber(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"42", 42},
		{"0", 0},
		{"123", 123},
		{"010", 8},   // octal
		{"077", 63},  // octal 77 = decimal 63
		{"0123", 83}, // octal 123 = decimal 83
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := lexNumber(t, tt.input)
			if got != tt.want {
				t.Errorf("Number() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLexerString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", `"hello"`, "hello"},
		{"empty", `""`, ""},
		{"with_spaces", `"hello world"`, "hello world"},
		{"with_escape_n", `"hello*nworld"`, "hello\nworld"},
		{"with_escape_t", `"tab*there"`, "tab\there"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexStringLiteral(t, tt.input)
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexerCharacter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{"single", `'a'`, int64('a')},
		{"multi", `'ab'`, int64('a')<<8 | int64('b')},
		{"escape_n", `'*n'`, int64('\n')},
		{"escape_t", `'*t'`, int64('\t')},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexCharacterLiteral(t, tt.input)
			if got != tt.want {
				t.Errorf("Character() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLexerWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  rune
	}{
		{"spaces", "   hello", 'h'},
		{"tabs", "\t\thello", 'h'},
		{"newlines", "\n\nhello", 'h'},
		{"mixed", " \t\n hello", 'h'},
		{"with_comment", "/* comment */ hello", 'h'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexWhitespaceNextRune(t, tt.input)
			if got != tt.want {
				t.Errorf("After Whitespace() got %c, want %c", got, tt.want)
			}
		})
	}
}

func TestLexerComment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "/* comment */ rest", " rest"},
		{"multiline", "/* line1\nline2 */ rest", " rest"},
		{"nested_stars", "/* ** comment ** */ rest", " rest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexCommentRest(t, tt.input)
			if got != tt.want {
				t.Errorf("After Comment() got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexerIdentifierWithWhitespaceAndComment(t *testing.T) {
	got := lexIdentifier(t, " \t /*c*/ name")
	if got != "name" {
		t.Errorf("Identifier() = %q, want %q", got, "name")
	}
}

func TestLexerEscapeUndefined(t *testing.T) {
	l := newTestLexer(t, "x")
	if _, err := l.Escape(); err == nil {
		t.Fatalf("Escape() expected error, got nil")
	}
}

func TestLexerUnterminatedString(t *testing.T) {
	l := newTestLexer(t, `"unterminated`)
	// Skip opening quote
	l.ReadChar()
	if _, err := l.String(); err == nil {
		t.Fatalf("String() expected error, got nil")
	}
}

func TestLexerUnclosedCharacter(t *testing.T) {
	l := newTestLexer(t, `'a`)
	// Skip opening quote
	l.ReadChar()
	if _, err := l.Character(); err == nil {
		t.Fatalf("Character() expected error, got nil")
	}
}

func TestLexerUnclosedComment(t *testing.T) {
	l := newTestLexer(t, "/* unterminated")
	// Skip opening /*
	l.ReadChar()
	l.ReadChar()
	if err := l.Comment(); err == nil {
		t.Fatalf("Comment() expected error, got nil")
	}
}

func TestLexerWhitespaceSlashEOF(t *testing.T) {
	got := lexWhitespaceNextRune(t, "/")
	if got != '/' {
		t.Errorf("After Whitespace() got %c, want %c", got, '/')
	}
}
