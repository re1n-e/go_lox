package main

import "fmt"

func Compile(source string) {
	var scanner Scanner
	scanner.InitScanner(source)
	line := -1
	for {
		token := scanner.scanToken()
		if token.line != line {
			fmt.Printf("%4d ", token.line)
			line = token.line
		} else {
			fmt.Printf("   | ")
		}

		// Properly slice and convert to string
		lexeme := string(token.start)
		fmt.Printf("%2d '%s'\n", token.typ_e, lexeme)

		if token.typ_e == TOKEN_EOF {
			break
		}
	}
}
