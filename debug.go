package main

import "fmt"

func (chunk *Chunk) DisassembleChunk(name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < len(chunk.Code); {
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
	case OP_NIL:
		return simpleInstruction("OP_NIL", offset)
	case OP_TRUE:
		return simpleInstruction("OP_TRUE", offset)
	case OP_FALSE:
		return simpleInstruction("OP_FALSE", offset)
	case OP_POP:
		return simpleInstruction("OP_POP", offset)
	case OP_GET_LOCAL:
		return chunk.byteInstruction("OP_GET_LOCAL", offset)
	case OP_SET_LOCAL:
		return chunk.byteInstruction("OP_SET_LOCAL", offset)
	case OP_GET_GLOBAL:
		return chunk.constantInstruction("OP_GET_GLOBAL", offset)
	case OP_DEFINE_GLOBAL:
		return chunk.constantInstruction("OP_DEFINE_GLOBAL", offset)
	case OP_SET_GLOBAL:
		return chunk.constantInstruction("OP_SET_GLOBAL", offset)
	case OP_EQUAL:
		return simpleInstruction("OP_EQUAL", offset)
	case OP_GREATER:
		return simpleInstruction("OP_GREATER", offset)
	case OP_LESS:
		return simpleInstruction("OP_LESS", offset)
	case OP_ADD:
		return simpleInstruction("OP_ADD", offset)
	case OP_SUBTRACT:
		return simpleInstruction("OP_SUBTRACT", offset)
	case OP_MULTIPLY:
		return simpleInstruction("OP_MULTIPLY", offset)
	case OP_DIVIDE:
		return simpleInstruction("OP_DIVIDE", offset)
	case OP_NOT:
		return simpleInstruction("OP_NOT", offset)
	case OP_NEGATE:
		return simpleInstruction("OP_NEGATE", offset)
	case OP_PRINT:
		return simpleInstruction("OP_PRINT", offset)
	case OP_JUMP:
		return chunk.jumpInstruction("OP_JUMP", 1, offset)
	case OP_JUMP_IF_FALSE:
		return chunk.jumpInstruction("OP_JUMP_IF_FALSE", 1, offset)
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

func (chunk *Chunk) byteInstruction(name string, offset int) int {
	slot := chunk.Code[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot)
	return offset + 2
}

func (chunk *Chunk) jumpInstruction(name string, sign int, offset int) int {
	jump := uint16(chunk.Code[offset+1]) << 8
	jump |= uint16(chunk.Code[offset+2])
	fmt.Printf("%-16s %4d -> %d\n", name, offset,
		offset+3+sign*int(jump))
	return offset + 3
}

func PrintValue(value Value) {
	switch value.Type {
	case VAL_BOOL:
		if value.Bool {
			fmt.Print("true")
		} else {
			fmt.Print("false")
		}
	case VAL_NIL:
		fmt.Print("nil")
	case VAL_NUMBER:
		fmt.Printf("%g", value.Num)
	case VAL_STRING:
		fmt.Printf("\"%s\"", value.String)
	}
}

func printValues(value Value) {
	PrintValue(value)
}
