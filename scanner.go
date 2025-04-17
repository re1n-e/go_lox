package main

import (
	"fmt"
	"unicode"
)

type TokenType int

// Single-character tokens.
const (
	TOKEN_LEFT_PAREN = iota
	TOKEN_RIGHT_PAREN
	TOKEN_LEFT_BRACE
	TOKEN_RIGHT_BRACE
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_MINUS
	TOKEN_PLUS
	TOKEN_SEMICOLON
	TOKEN_SLASH
	TOKEN_STAR
	// One or two character tokens.
	TOKEN_BANG
	TOKEN_BANG_EQUAL
	TOKEN_EQUAL
	TOKEN_EQUAL_EQUAL
	TOKEN_GREATER
	TOKEN_GREATER_EQUAL
	TOKEN_LESS
	TOKEN_LESS_EQUAL
	// Literals.
	TOKEN_IDENTIFIER
	TOKEN_STRING
	TOKEN_NUMBER
	// Keywords.
	TOKEN_AND
	TOKEN_CLASS
	TOKEN_ELSE
	TOKEN_FALSE
	TOKEN_FOR
	TOKEN_FUN
	TOKEN_IF
	TOKEN_NIL
	TOKEN_OR
	TOKEN_PRINT
	TOKEN_RETURN
	TOKEN_SUPER
	TOKEN_THIS
	TOKEN_TRUE
	TOKEN_VAR
	TOKEN_WHILE

	TOKEN_ERROR
	TOKEN_EOF
)

type Scanner struct {
	Source  []rune
	Start   int
	Current int
	Line    int
}

func (scanner *Scanner) InitScanner(source string) {
	scanner.Source = []rune(source)
	scanner.Current = 0
	scanner.Line = 1
}

type Token struct {
	typ_e  TokenType
	start  []rune
	length int
	line   int
}

func (scanner *Scanner) scanToken() Token {
	scanner.skipWhitespace()
	scanner.Start = scanner.Current

	if scanner.isAtEnd() {
		return scanner.makeToken(TOKEN_EOF)
	}

	c := scanner.advance()
	if unicode.IsLetter(c) {
		return scanner.identifier()
	}
	if unicode.IsDigit(c) {
		return scanner.number()
	}

	switch c {
	case '(':
		return scanner.makeToken(TOKEN_LEFT_PAREN)
	case ')':
		return scanner.makeToken(TOKEN_RIGHT_PAREN)
	case '{':
		return scanner.makeToken(TOKEN_LEFT_BRACE)
	case '}':
		return scanner.makeToken(TOKEN_RIGHT_BRACE)
	case ';':
		return scanner.makeToken(TOKEN_SEMICOLON)
	case ',':
		return scanner.makeToken(TOKEN_COMMA)
	case '.':
		return scanner.makeToken(TOKEN_DOT)
	case '-':
		return scanner.makeToken(TOKEN_MINUS)
	case '+':
		return scanner.makeToken(TOKEN_PLUS)
	case '/':
		return scanner.makeToken(TOKEN_SLASH)
	case '*':
		return scanner.makeToken(TOKEN_STAR)
	case '!':
		if scanner.match('=') {
			return scanner.makeToken(TOKEN_BANG_EQUAL)
		}
		return scanner.makeToken(TOKEN_BANG)
	case '=':
		if scanner.match('=') {
			return scanner.makeToken(TOKEN_EQUAL_EQUAL)
		}
		return scanner.makeToken(TOKEN_EQUAL)
	case '<':
		if scanner.match('=') {
			return scanner.makeToken(TOKEN_LESS_EQUAL)
		}
		return scanner.makeToken(TOKEN_LESS)
	case '>':
		if scanner.match('=') {
			return scanner.makeToken(TOKEN_GREATER_EQUAL)
		}
		return scanner.makeToken(TOKEN_GREATER)
	case '"':
		return scanner.string()
	}
	return scanner.errorToken(fmt.Sprintf("Unexpected character. %c", c))
}

func (scanner *Scanner) advance() rune {
	if scanner.isAtEnd() {
		return 0
	}
	scanner.Current++
	return scanner.Source[scanner.Current-1]
}

func (scanner *Scanner) peek() rune {
	if scanner.isAtEnd() {
		return 0
	}
	return scanner.Source[scanner.Current]
}

func (scanner *Scanner) peekNext() rune {
	if scanner.isAtEnd() {
		return ' '
	}
	return scanner.Source[scanner.Current+1]
}

func (scanner *Scanner) match(expected rune) bool {
	if scanner.isAtEnd() {
		return false
	}
	if scanner.Source[scanner.Current] != expected {
		return false
	}
	scanner.Current++
	return true
}

func (scanner *Scanner) makeToken(typ_e TokenType) Token {
	var token Token
	token.typ_e = typ_e
	token.start = scanner.Source[scanner.Start:scanner.Current]
	token.length = scanner.Current - scanner.Start
	token.line = scanner.Line
	return token
}

func (scanner *Scanner) errorToken(message string) Token {
	var token Token
	token.typ_e = TOKEN_ERROR
	token.start = []rune(message)
	token.length = len(message)
	token.line = scanner.Line
	return token
}

func (scanner *Scanner) skipWhitespace() {
	for !scanner.isAtEnd() {
		switch scanner.peek() {
		case ' ', '\r', '\t', '\n':
			if scanner.peek() == '\n' {
				scanner.Line++
			}
			scanner.advance()
		case '/':
			if scanner.peekNext() == '/' {
				for scanner.peek() != '\n' && !scanner.isAtEnd() {
					scanner.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}
func (scanner *Scanner) checkKeyword(start int, length int, rest string, typ_e TokenType) TokenType {
	if scanner.Current-scanner.Start == start+length {
		slice := string(scanner.Source[scanner.Start+start : scanner.Start+start+length])
		if slice == rest {
			return typ_e
		}
	}
	return TOKEN_IDENTIFIER
}

func (scanner *Scanner) identifier() Token {
	for unicode.IsLetter(scanner.peek()) || unicode.IsDigit(scanner.peek()) {
		scanner.advance()
	}
	return scanner.makeToken(scanner.identifierType())
}

func (scanner *Scanner) identifierType() TokenType {
	switch scanner.Source[scanner.Start] {
	case 'a':
		return scanner.checkKeyword(1, 2, "nd", TOKEN_AND)
	case 'c':
		return scanner.checkKeyword(1, 4, "lass", TOKEN_CLASS)
	case 'e':
		return scanner.checkKeyword(1, 3, "lse", TOKEN_ELSE)
	case 'f':
		if scanner.Current-scanner.Start > 1 {
			switch scanner.Source[scanner.Start+1] {
			case 'a':
				return scanner.checkKeyword(1, 4, "alse", TOKEN_FALSE)
			case 'o':
				return scanner.checkKeyword(1, 2, "or", TOKEN_FOR)
			case 'u':
				return scanner.checkKeyword(1, 2, "un", TOKEN_FUN)
			}
		}
	case 'i':
		return scanner.checkKeyword(1, 1, "f", TOKEN_IF)
	case 'n':
		return scanner.checkKeyword(1, 2, "il", TOKEN_NIL)
	case 'o':
		return scanner.checkKeyword(1, 1, "r", TOKEN_OR)
	case 'p':
		return scanner.checkKeyword(1, 4, "rint", TOKEN_PRINT)
	case 'r':
		return scanner.checkKeyword(1, 5, "eturn", TOKEN_RETURN)
	case 's':
		return scanner.checkKeyword(1, 4, "uper", TOKEN_SUPER)
	case 't':
		if scanner.Current-scanner.Start > 1 {
			switch scanner.Source[scanner.Start+1] {
			case 'h':
				return scanner.checkKeyword(1, 3, "his", TOKEN_THIS)
			case 'r':
				return scanner.checkKeyword(1, 3, "rue", TOKEN_TRUE)
			}
		}
	case 'v':
		return scanner.checkKeyword(1, 2, "ar", TOKEN_VAR)
	case 'w':
		return scanner.checkKeyword(1, 4, "hile", TOKEN_WHILE)
	}
	return TOKEN_IDENTIFIER
}

func (scanner *Scanner) number() Token {
	for unicode.IsDigit(scanner.peek()) {
		scanner.advance()
	}

	if scanner.peek() == '.' && unicode.IsDigit(scanner.peekNext()) {
		scanner.advance()

		for unicode.IsDigit(scanner.peek()) {
			scanner.advance()
		}
	}

	return scanner.makeToken(TOKEN_NUMBER)
}

func (scanner *Scanner) string() Token {
	for scanner.peek() != '"' && !scanner.isAtEnd() {
		if scanner.peek() == '\n' {
			scanner.Line++
		}
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return scanner.errorToken("Unterminated string.")
	}

	scanner.advance()
	return scanner.makeToken(TOKEN_STRING)
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.Current == len(scanner.Source)
}
