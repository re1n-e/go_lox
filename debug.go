package main

import "fmt"

func (chunk *Chunk) DisassembleChunk(name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < chunk.Count; {
		offset = chunk.disassembleInstruction(offset)
	}
}

func (chunk *Chunk) disassembleInstruction(offset int) int {
	if offset >= len(chunk.Lines) {
		fmt.Printf("Error: no line info for offset %d\n", offset)
		return offset + 1
	}
	fmt.Printf("%04d ", offset)
	if offset > 0 && chunk.Lines[offset] == chunk.Lines[offset-1] {
		fmt.Printf("   | ")
	} else {
		fmt.Printf("%4d ", chunk.Lines[offset])
	}
	instruction := chunk.Code[offset]
	switch instruction {
	case OP_CONSTANT:
		return chunk.constantInstruction("OP_CONSTANT", offset)
	case OP_ADD:
		return simpleInstruction("OP_ADD", offset)
	case OP_SUBTRACT:
		return simpleInstruction("OP_SUBTRACT", offset)
	case OP_MULTIPLY:
		return simpleInstruction("OP_MULTIPLY", offset)
	case OP_DIVIDE:
		return simpleInstruction("OP_DIVIDE", offset)
	case OP_NEGATE:
		return simpleInstruction("OP_NEGATE", offset)
	case OP_RETURN:
		return simpleInstruction("OP_RETURN", offset)
	default:
		fmt.Printf("Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

func (chunk *Chunk) constantInstruction(name string, offset int) int {
	if offset+1 >= len(chunk.Code) {
		fmt.Printf("Error: %s instruction at offset %d missing operand\n", name, offset)
		return offset + 1
	}

	constant := chunk.Code[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	if int(constant) >= len(chunk.Constants) {
		fmt.Printf("Error: constant index %d out of bounds\n", constant)
	} else {
		printValues(chunk.Constants[int(constant)])
	}
	fmt.Println("'")
	return offset + 2
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func printValues(value Value) {
	fmt.Printf("%g", value)
}
