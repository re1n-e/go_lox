package main

type Parser struct {
	current  Token
	previous Token
}

func Compile(source string, chunk *Chunk) bool {
	var parser Parser
	var scanner Scanner
	scanner.InitScanner(source)
	parser.advance(&scanner)
	return true
}

func (parser *Parser) advance(scanner *Scanner) {
	parser.previous = parser.current
	for {
		parser.current = scanner.scanToken()
		if parser.current.typ_e != TOKEN_ERROR {
			break
		}
	}
}

func consume()
