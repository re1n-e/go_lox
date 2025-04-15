package main

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
	scanner.Current = 1
	scanner.Line = 1
}

type Token struct {
	typ_e  TokenType
	start  []rune
	length int
	line   int
}

func (scanner *Scanner) scanToken() Token {
	scanner.Start = scanner.Current

	if scanner.isAtEnd() {
		return scanner.makeToken(TOKEN_EOF)
	}

	return scanner.errorToken("Unexpected character.")
}

func (scanner *Scanner) makeToken(typ_e TokenType) Token {
	var token Token
	token.typ_e = typ_e
	token.start = scanner.Source
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

func (scanner *Scanner) isAtEnd() bool {
	return scanner.Current == len(scanner.Source)
}
