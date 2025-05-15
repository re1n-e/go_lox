package main

// currentChunk = len(compilingChunk.Code)

import (
	"fmt"
	"os"
	"strconv"
)

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

type Local struct {
	name  Token
	depth int
}

type Compiler struct {
	locals     []Local
	localCount int
	scopeDepth int
}

var scanner *Scanner
var compilingChunk *Chunk
var current *Compiler

func compiler_init() {
	current = &Compiler{}
	current.locals = make([]Local, 256)
	current.localCount = 0
	current.scopeDepth = 0
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
		{nil, func(p *Parser, canAssign bool) { p.and_(canAssign) }, PREC_NONE}, // And
		{nil, nil, PREC_NONE}, // Class
		{func(p *Parser, canAssign bool) { p.literal(canAssign) }, nil, PREC_NONE}, // Else
		{func(p *Parser, canAssign bool) { p.literal(canAssign) }, nil, PREC_NONE}, // False
		{nil, nil, PREC_NONE}, // For
		{nil, nil, PREC_NONE}, // Fun
		{nil, nil, PREC_NONE}, // If
		{func(p *Parser, canAssign bool) { p.literal(canAssign) }, nil, PREC_NONE}, // NIL
		{nil, func(p *Parser, canAssign bool) { p.or_(canAssign) }, PREC_NONE},     // OR
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
	compiler_init()
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

func (parser *Parser) emitJump(instruction byte) byte {
	parser.emitByte(instruction)
	parser.emitByte(0xff)
	parser.emitByte(0xff)
	return byte(len(compilingChunk.Code)) - 2
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

func (parser *Parser) patchJump(offset byte) {
	jump := len(compilingChunk.Code) - int(offset) - 2

	if jump > 255 {
		parser.error("Too much code to jump over")
	}

	compilingChunk.Code[offset] = byte((jump >> 8) & 0xff)
	compilingChunk.Code[offset+1] = byte(jump & 0xff)
}

func (parser *Parser) emitBytes(byte1, byte2 byte) {
	parser.emitByte(byte1)
	parser.emitByte(byte2)
}

func (parser *Parser) emitLoop(loopStart int) {
	parser.emitByte(OP_LOOP)

	offset := len(compilingChunk.Code) - loopStart + 2
	if offset > int(^uint16(0)) {
		parser.error("Loop body too large.")
	}

	parser.emitByte((byte(offset >> 8)) & 0xff)
	parser.emitByte(byte(offset) & 0xff)
}

func (parser *Parser) endCompiler() {
	parser.emitReturn()
	if !parser.hadError {
		// compilingChunk.DisassembleChunk("code")
	}
}

func beginScope() {
	current.scopeDepth++
}

func (parser *Parser) endScope() {
	for current.localCount > 0 &&
		current.locals[current.localCount-1].depth > current.scopeDepth {
		parser.emitByte(OP_POP)
		current.localCount--
	}

	current.scopeDepth--
}

func (parser *Parser) binary(bool) {
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

func (parser *Parser) literal(bool) {
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

func (parser *Parser) grouping(bool) {
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression")
}

func (parser *Parser) number(bool) {
	value, err := strconv.ParseFloat(string(parser.previous.start), 64)
	if err != nil {
		panic("number() cant't convert")
	}
	parser.emitConstant(NumberVal(value))
}

func (parser *Parser) or_(bool) {
	elseJump := parser.emitJump(OP_JUMP_IF_FALSE)
	endJump := parser.emitJump(OP_JUMP)

	parser.patchJump(elseJump)
	parser.emitByte(OP_POP)

	parser.parsePrecedence(PREC_OR)
	parser.patchJump(endJump)
}

func (parser *Parser) string(bool) {
	value := string(parser.previous.start[1 : len(parser.previous.start)-1])
	parser.emitConstant(StringVal(value))
}

func (parser *Parser) namedVariable(name Token, canAssign bool) {
	var getOp, setOp uint8
	get_arg := parser.resolveLocal(current, name)
	var arg byte
	if get_arg != -1 {
		arg = byte(get_arg)
		getOp = OP_GET_LOCAL
		setOp = OP_SET_LOCAL
	} else {
		arg = parser.identifierConstant(name)
		getOp = OP_GET_GLOBAL
		setOp = OP_SET_GLOBAL
	}

	if canAssign && parser.match(TOKEN_EQUAL) {
		parser.expression()
		parser.emitBytes(setOp, arg)
	} else {
		parser.emitBytes(getOp, arg)
	}
}

func (parser *Parser) variable(canAssign bool) {
	parser.namedVariable(parser.previous, canAssign)
}

func (parser *Parser) unary(bool) {
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

func identifiersEqual(a, b *Token) bool {
	if a.length != b.length {
		return false
	}
	for i := 0; i < a.length; i++ {
		if a.start[i] != b.start[i] {
			return false
		}
	}
	return true
}

func (parser *Parser) resolveLocal(compiler *Compiler, name Token) int {
	for i := compiler.localCount - 1; i >= 0; i-- {
		local := &compiler.locals[i]
		if identifiersEqual(&name, &local.name) {
			if local.depth == -1 {
				parser.error("Can't read local variable in its own initializer.")
			}
			return i
		}
	}
	return -1
}

func (parser *Parser) declareVariable() {
	if current.scopeDepth == 0 {
		return
	}

	name := parser.previous
	for i := current.localCount - 1; i >= 0; i-- {
		local := &current.locals[i]
		if local.depth != -1 && local.depth < current.scopeDepth {
			break
		}

		if identifiersEqual(&name, &local.name) {
			parser.error("Already a variable with this name in this scope.")
		}
	}
	parser.addLocal(name)
}

func (parser *Parser) addLocal(name Token) {
	if current.localCount == 256 {
		parser.error("Too many local variable in function.")
		return
	}
	local := &current.locals[current.localCount]
	current.localCount++
	local.name = name
	local.depth = -1
	local.depth = current.scopeDepth
}

func (parser *Parser) parseVariable(errorMessage string) byte {
	parser.consume(TOKEN_IDENTIFIER, errorMessage)

	parser.declareVariable()
	if current.scopeDepth > 0 {
		return 0
	}

	return parser.identifierConstant(parser.previous)
}

func markInitialized() {
	current.locals[current.localCount-1].depth = current.scopeDepth
}

func (parser *Parser) defineVariable(global byte) {
	if current.scopeDepth > 0 {
		markInitialized()
		return
	}
	parser.emitBytes(OP_DEFINE_GLOBAL, global)
}

func (parser *Parser) and_(bool) {
	endJump := parser.emitJump(OP_JUMP_IF_FALSE)
	parser.emitByte(OP_POP)
	parser.parsePrecedence(PREC_AND)

	parser.patchJump(endJump)
}

func getRule(Type TokenType) *ParseRule {
	return &rules[Type]
}

func (parser *Parser) expression() {
	parser.parsePrecedence(PREC_ASSIGNMENT)
}

func (parser *Parser) block() {
	for !parser.check(TOKEN_RIGHT_BRACE) && !parser.check(TOKEN_EOF) {
		parser.declaration()
	}

	parser.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
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

func (parser *Parser) ifStatement() {
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'if'.")
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	thenJump := parser.emitJump(OP_JUMP_IF_FALSE)
	parser.emitByte(OP_POP)
	parser.statement()

	elseJump := parser.emitJump(OP_JUMP)

	parser.patchJump(thenJump)
	parser.emitByte(OP_POP)

	if parser.match(TOKEN_ELSE) {
		parser.statement()
	}
	parser.patchJump(elseJump)
}

func (parser *Parser) printStatement() {
	parser.expression()
	parser.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	parser.emitByte(OP_PRINT)
}

func (parser *Parser) whileStatement() {
	loopStart := len(compilingChunk.Code)
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'while'.")
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := parser.emitJump(OP_JUMP_IF_FALSE)
	parser.emitByte(OP_POP)
	parser.statement()
	parser.emitLoop(loopStart)

	parser.patchJump(exitJump)
	parser.emitByte(OP_POP)
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
	} else if parser.match(TOKEN_FOR){
		parser.forStatement()
	}else if parser.match(TOKEN_IF) {
		parser.ifStatement()
	} else if parser.match(TOKEN_WHILE) {
		parser.whileStatement()
	} else if parser.match(TOKEN_LEFT_BRACE) {
		beginScope()
		parser.block()
		parser.endScope()
	} else {
		parser.expressionStatement()
	}
}
