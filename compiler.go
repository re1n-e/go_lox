package main

import (
	"fmt"
	"os"
	"strconv"
)

var scanner Scanner
var compilingChunk *Chunk

type Parser struct {
	current   Token
	previous  Token
	panicMode bool
	hadError  bool
}

const (
	PREC_NONE       = iota
	PREC_ASSIGNMENT // =
	PREC_OR         // or
	PREC_AND        // and
	PREC_EQUALITY   // == !=
	PREC_COMPARISON // < > <= >=
	PREC_TERM       // + -
	PREC_FACTOR     // * /
	PREC_UNARY      // ! -
	PREC_CALL       // . ()
	PREC_PRIMARY
)

type Precedence int

type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence Precedence
}

type ParseFn func()

func Compile(source string, chunk *Chunk) bool {
	var parser Parser

	scanner.InitScanner(source)
	compilingChunk = chunk
	parser.hadError = false
	parser.panicMode = false

	parser.advance()
	parser.expression()

	parser.consume(TOKEN_EOF, "Expect end of expression.")
	parser.endCompiler()
	return !parser.hadError
}

func (parser *Parser) advance() {
	parser.previous = parser.current
	for {
		parser.current = scanner.scanToken()
		if parser.current.typ_e != TOKEN_ERROR {
			break
		}
	}
}

func (parser *Parser) errorAtCurrent(message string) {
	parser.errorAt(&parser.previous, message)
}

func (parser *Parser) error(message string) {
	parser.errorAt(&parser.previous, message)
}

func (parser *Parser) errorAt(token *Token, message string) {
	if parser.panicMode {
		return
	}
	parser.panicMode = true
	fmt.Fprintf(os.Stderr, "[line %d] Error", token.line)
	if token.typ_e == TOKEN_EOF {
		fmt.Fprintf(os.Stderr, " at end")
	} else if token.typ_e == TOKEN_ERROR {
		// Nothing
	} else {
		fmt.Fprintf(os.Stderr, " at '%s'", string(token.start))
	}

	fmt.Fprintf(os.Stderr, ": %s", message)
	parser.hadError = true
}

func (parser *Parser) consume(typ_e TokenType, message string) {
	if parser.current.typ_e == typ_e {
		parser.advance()
		return
	}

	parser.errorAtCurrent(message)
}

func (parser *Parser) emitByte(Byte byte) {
	compilingChunk.WriteChunk(Byte, parser.previous.line)
}

func (parser *Parser) emitReturn() {
	parser.emitByte(OP_RETURN)
}

func (parser *Parser) makeConstant(value Value) byte {
	constant := compilingChunk.AddConstant(value)
	if constant > 255 {
		parser.error("Too many constants in one chunk.")
		return 0
	}
	return byte(constant)
}

func (parser *Parser) emitConstant(value Value) {
	parser.emitBytes(OP_CONSTANT, parser.makeConstant(value))
}

func (parser *Parser) emitBytes(byte1, byte2 byte) {
	parser.emitByte(byte1)
	parser.emitByte(byte2)
}

func (parser *Parser) endCompiler() {
	parser.emitReturn()
}

func (parser *Parser) binary() {
	operatorType := parser.previous.typ_e

}

func (parser *Parser) grouping() {
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression")
}

func (parser *Parser) number() {
	value, err := strconv.ParseFloat(string(parser.previous.start), 64)
	if err != nil {
		panic("number() cant't convert")
	}
	parser.emitConstant(Value(value))
}

func (parser *Parser) unary() {
	operatorType := parser.previous.typ_e
	// compile the operand
	parser.parsePrecedence(PREC_UNARY)

	// Emit the operator instruction
	switch operatorType {
	case TOKEN_MINUS:
		parser.emitByte(OP_NEGATE)
	default:
		return
	}
}

func (parser *Parser) parsePrecedence(precdence Precedence) {

}

func (parser *Parser) expression() {
	parser.parsePrecedence(PREC_ASSIGNMENT)
}
