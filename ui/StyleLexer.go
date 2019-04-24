package ui

import (
	"unicode"

	"github.com/eczarny/lexer"
)

// These are our lexer token values.
const (
	TokenProperty lexer.TokenType = iota
	TokenValue
	TokenNewline

	TokenComment

	TokenEOF
)

func getTokenName(token lexer.Token) string {
	switch token.Type {
	case TokenProperty:
		return "PROP"
	case TokenValue:
		return "VAL"
	case TokenNewline:
		return "NL"
	case TokenComment:
		return "CMT"
	case TokenEOF:
		return "EOF"
	}
	return "Err"
}

// NewObjectLexer creates a new lexer from the passed string.
func NewObjectLexer(input string) *lexer.Lexer {
	return lexer.NewLexer(input, initialState)
}

// states

func initialState(l *lexer.Lexer) lexer.StateFunc {
	r := l.IgnoreUpTo(func(r rune) bool {
		return tNewline(r) || tNonWhitespace(r)
	})
	switch {
	case tComment(r):
		return commentState
	case tNewline(r):
		return newlineState
	case tNonWhitespace(r) && r != lexer.EOF:
		return variableState
	}
	l.Emit(TokenEOF)
	return nil
}

func commentState(l *lexer.Lexer) lexer.StateFunc {
	l.IgnoreUpTo(func(r rune) bool {
		return tNewline(r)
	})
	return initialState
}

func valueState(l *lexer.Lexer) lexer.StateFunc {
	l.IgnoreUpTo(func(r rune) bool {
		return tNonWhitespace(r)
	})
	// If the next rune is NL or ';', presume empty Value
	nr := l.Peek()
	if tNewline(nr) || tComment(nr) {
		return initialState
	}
	l.NextUpTo(func(r rune) bool {
		return tNewline(r) || tComment(r)
	})
	l.Emit(TokenValue)
	return initialState
}

func variableState(l *lexer.Lexer) lexer.StateFunc {
	r := l.NextUpTo(func(r rune) bool {
		return tWhitespace(r) || tComment(r)
	})
	switch {
	case tComment(r):
		l.Emit(TokenProperty)
		return commentState
	case tNewline(r):
		l.Emit(TokenProperty)
		return initialState
	case tWhitespace(r):
		l.Emit(TokenProperty)
		return valueState
	}
	return initialState
}

func newlineState(l *lexer.Lexer) lexer.StateFunc {
	l.Ignore()
	return initialState
}

// type detection

func tWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func tNonWhitespace(r rune) bool {
	return !tWhitespace(r)
}

func tAlphanumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func tNonAlphanumeric(r rune) bool {
	return !tAlphanumeric(r)
}

func tComment(r rune) bool {
	return r == ';'
}

func tNewline(r rune) bool {
	return r == '\n' || r == '\r'
}
