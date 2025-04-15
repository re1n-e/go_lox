package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	var c Chunk
	var v VM
	c.InitChunk()
	v.InitVM()

	if len(os.Args) == 1 {
		repl(&v)
	} else if len(os.Args) == 2 {
		runFile(&v, os.Args[1])
	} else {
		fmt.Fprintf(os.Stderr, "Usage: ./main [path]\n")
		os.Exit(64)
	}
}

func repl(v *VM) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			break
		}
		v.Interpret(line)
	}
}

func runFile(v *VM, path string) {
	source := readFile(path)
	result := v.Interpret(source)

	switch result {
	case INTERPRET_COMPILE_ERROR:
		os.Exit(65)
	case INTERPRET_RUNTIME_ERROR:
		os.Exit(70)
	}
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file: %s\n", err)
		os.Exit(74)
	}
	return string(data)
}
