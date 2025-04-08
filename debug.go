package main

import "fmt"

func (chunk *Chunk) DisassembleChunk(name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < chunk.Count; {
		offset = chunk.DisassembleInstruction(offset)
	}
}

func (chunk *Chunk) DisassembleInstruction(offset int) int {

}