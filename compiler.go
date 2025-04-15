package main

func Compile(source string) {
	var scanner Scanner
	scanner.InitScanner(source)
	line := -1
	for {
		token := scanner.scanToken()
	}
}
