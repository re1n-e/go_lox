package main

import (
	"fmt"
	"os"
	"strconv"
)

var scanner *Scanner
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

type ParseFn func(*Parser, bool)

type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence Precedence
}

var rules []ParseRule

func rules_init() {
	rules = []ParseRule{
		{
			Prefix:     func(p *Parser, canAssign bool) { p.grouping(canAssign) }, //Left Paren
			Infix:      nil,
			Precedence: PREC_NONE,
		},
		{nil, nil, PREC_NONE}, // Right Paren
		{nil, nil, PREC_NONE}, // Left Brace
		{nil, nil, PREC_NONE}, // Right Brace
		{nil, nil, PREC_NONE}, // Comma
		{nil, nil, PREC_NONE}, // Dot
		{
			Prefix:     func(p *Parser, canAssign bool) { p.unary(canAssign) },
			Infix:      func(p *Parser, canAssign bool) { p.binary(canAssign) }, // Minus
			Precedence: PREC_TERM,
		},
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_TERM}, // plus
		{nil, nil, PREC_NONE}, // Semicolon
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_FACTOR},   // slash
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_FACTOR},   // star
		{func(p *Parser, canAssign bool) { p.unary(canAssign) }, nil, PREC_NONE},      // bang
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_EQUALITY}, // bang equal
		{nil, nil, PREC_NONE}, // Equal
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_EQUALITY},   // Equal Equal
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_COMPARISON}, // Greater
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_COMPARISON}, // Greater Equal
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_COMPARISON}, // Less
		{nil, func(p *Parser, canAssign bool) { p.binary(canAssign) }, PREC_COMPARISON}, // Less Equal
		{func(p *Parser, canAssign bool) { p.variable(canAssign) }, nil, PREC_NONE},     // Identifier
		{func(p *Parser, canAssign bool) { p.string(canAssign) }, nil, PREC_NONE},       // String
		{
			Prefix:     func(p *Parser, canAssign bool) { p.number(canAssign) }, // Number
			Infix:      nil,
			Precedence: PREC_NONE,
		},
		{nil, nil, PREC_NONE}, // And
		{nil, nil, PREC_NONE}, // Class
		{func(p *Parser, canAssign bool) { p.literal(canAssign) }, nil, PREC_NONE}, // Else
		{func(p *Parser, canAssign bool) { p.literal(canAssign) }, nil, PREC_NONE}, // False
		{nil, nil, PREC_NONE}, // For
		{nil, nil, PREC_NONE}, // Fun
		{nil, nil, PREC_NONE}, // If
		{func(p *Parser, canAssign bool) { p.literal(canAssign) }, nil, PREC_NONE}, // NIL
		{nil, nil, PREC_NONE}, // OR
		{nil, nil, PREC_NONE}, // Print
		{nil, nil, PREC_NONE}, // Return
		{nil, nil, PREC_NONE}, // Super
		{nil, nil, PREC_NONE}, // This
		{func(p *Parser, canAssign bool) { p.literal(canAssign) }, nil, PREC_NONE}, // True
		{nil, nil, PREC_NONE}, // Var
		{nil, nil, PREC_NONE}, // While
		{nil, nil, PREC_NONE}, // Error
		{nil, nil, PREC_NONE}, // Eof
	}
}

func Compile(source string, chunk *Chunk) bool {
	var parser Parser
	scanner = &Scanner{}

	scanner.InitScanner(source)
	compilingChunk = chunk
	parser.hadError = false
	parser.panicMode = false

	parser.advance()
	rules_init()

	for !parser.match(TOKEN_EOF) {
		parser.declaration()
	}

	parser.endCompiler()
	return !parser.hadError
}

func (parser *Parser) advance() {
	parser.previous = parser.current
	for {
		parser.current = scanner.scanToken()
		if parser.current.Type != TOKEN_ERROR {
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
	if token.Type == TOKEN_EOF {
		fmt.Fprintf(os.Stderr, " at end")
	} else if token.Type == TOKEN_ERROR {
		// Nothing
	} else {
		fmt.Fprintf(os.Stderr, " at '%s'", string(token.start))
	}
	fmt.Fprintf(os.Stderr, ": %s", message)
	fmt.Println()
	parser.hadError = true
}

func (parser *Parser) consume(Type TokenType, message string) {
	if parser.current.Type == Type {
		parser.advance()
		return
	}

	parser.errorAtCurrent(message)
}

func (parser *Parser) check(Type TokenType) bool {
	return parser.current.Type == Type
}

func (parser *Parser) match(Type TokenType) bool {
	if !parser.check(Type) {
		return false
	}
	parser.advance()
	return true
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
	if !parser.hadError {
		compilingChunk.DisassembleChunk("code")
	}
}

func (parser *Parser) binary(canAssign bool) {
	operatorType := parser.previous.Type
	rule := getRule(operatorType)

	parser.parsePrecedence(Precedence(rule.Precedence + 1))

	switch operatorType {
	case TOKEN_BANG_EQUAL:
		parser.emitBytes(OP_EQUAL, OP_NOT)
	case TOKEN_EQUAL_EQUAL:
		parser.emitByte(OP_EQUAL)
	case TOKEN_GREATER:
		parser.emitByte(OP_GREATER)
	case TOKEN_GREATER_EQUAL:
		parser.emitBytes(OP_LESS, OP_NOT)
	case TOKEN_LESS:
		parser.emitByte(OP_LESS)
	case TOKEN_LESS_EQUAL:
		parser.emitBytes(OP_GREATER, OP_NOT)
	case TOKEN_PLUS:
		parser.emitByte(OP_ADD)
	case TOKEN_MINUS:
		parser.emitByte(OP_SUBTRACT)
	case TOKEN_STAR:
		parser.emitByte(OP_MULTIPLY)
	case TOKEN_SLASH:
		parser.emitByte(OP_DIVIDE)
	default:
		return
	}
}

func (parser *Parser) literal(canAssign bool) {
	switch parser.previous.Type {
	case TOKEN_FALSE:
		parser.emitByte(OP_FALSE)
	case TOKEN_NIL:
		parser.emitByte(OP_NIL)
	case TOKEN_TRUE:
		parser.emitByte(OP_TRUE)
	default:
		return
	}
}

func (parser *Parser) grouping(canAssign bool) {
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression")
}

func (parser *Parser) number(canAssign bool) {
	value, err := strconv.ParseFloat(string(parser.previous.start), 64)
	if err != nil {
		panic("number() cant't convert")
	}
	parser.emitConstant(NumberVal(value))
}

func (parser *Parser) string(canAssign bool) {
	value := string(parser.previous.start[1 : len(parser.previous.start)-1])
	parser.emitConstant(StringVal(value))
}

func (parser *Parser) namedVariable(name Token, canAssign bool) {
	arg := parser.identifierConstant(name)

	if canAssign && parser.match(TOKEN_EQUAL) {
		parser.expression()
		parser.emitBytes(OP_SET_GLOBAL, arg)
	} else {

		parser.emitBytes(OP_GET_GLOBAL, arg)
	}
}
func (parser *Parser) variable(canAssign bool) {
	parser.namedVariable(parser.previous, canAssign)
}

func (parser *Parser) unary(canAssign bool) {
	operatorType := parser.previous.Type
	// compile the operand
	parser.parsePrecedence(PREC_UNARY)

	// Emit the operator instruction
	switch operatorType {
	case TOKEN_BANG:
		parser.emitByte(OP_NOT)
	case TOKEN_MINUS:
		parser.emitByte(OP_NEGATE)
	default:
		return
	}
}

func (parser *Parser) parsePrecedence(precedence Precedence) {
	parser.advance()
	prefixRule := getRule(parser.previous.Type).Prefix
	if prefixRule == nil {
		parser.error("Expect expression.\n")
		return
	}

	canAssign := precedence <= PREC_ASSIGNMENT
	prefixRule(parser, canAssign)

	for precedence <= getRule(parser.current.Type).Precedence {
		parser.advance()
		infixRule := getRule(parser.previous.Type).Infix
		infixRule(parser, canAssign)
	}

	if canAssign && parser.match(TOKEN_EQUAL) {
		parser.error("Invalid assignment target.")
	}
}

func (parser *Parser) identifierConstant(name Token) byte {
	identifier := string(name.start)

	value := StringVal(identifier)

	return parser.makeConstant(value)
}

func (parser *Parser) parseVariable(errorMessage string) byte {
	parser.consume(TOKEN_IDENTIFIER, errorMessage)
	return parser.identifierConstant(parser.previous)
}

func (parser *Parser) defineVariable(global byte) {
	parser.emitBytes(OP_DEFINE_GLOBAL, global)
}

func getRule(Type TokenType) *ParseRule {
	return &rules[Type]
}

func (parser *Parser) expression() {
	parser.parsePrecedence(PREC_ASSIGNMENT)
}

func (parser *Parser) varDeclaration() {
	global := parser.parseVariable("Expect variable name.")
	if parser.match(TOKEN_EQUAL) {
		parser.expression()
	} else {
		parser.emitByte(OP_NIL)
	}

	parser.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration.")

	parser.defineVariable(global)
}

func (parser *Parser) expressionStatement() {
	parser.expression()
	parser.consume(TOKEN_SEMICOLON, "Expect ';' after expression.")
	parser.emitByte(OP_POP)
}

func (parser *Parser) printStatement() {
	parser.expression()
	parser.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	parser.emitByte(OP_PRINT)
}

func (parser *Parser) synchronize() {
	parser.panicMode = false

	for parser.current.Type != TOKEN_EOF {
		if parser.previous.Type == TOKEN_SEMICOLON {
			return
		}
		switch parser.current.Type {
		case TOKEN_CLASS,
			TOKEN_FUN,
			TOKEN_VAR,
			TOKEN_FOR,
			TOKEN_IF,
			TOKEN_WHILE,
			TOKEN_PRINT,
			TOKEN_RETURN:
			return
		default:

		}
	}

	parser.advance()
}

func (parser *Parser) declaration() {
	if parser.match(TOKEN_VAR) {
		parser.varDeclaration()
	} else {
		parser.statement()
	}

	if parser.panicMode {
		parser.synchronize()
	}
}

func (parser *Parser) statement() {
	if parser.match(TOKEN_PRINT) {
		parser.printStatement()
	} else {
		parser.expressionStatement()
	}
}
