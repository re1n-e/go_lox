package main

import (
	"fmt"
	"os"
	"unicode"
)

type InterpretResult int

const STACK_MAX = 256

const (
	INTERPRET_OK = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	Chunk       Chunk
	Instruction []uint8
	Ip          int
	Stack       []Value
	Sp          int
}

func (vm *VM) InitVM() {
	vm.Stack = make([]Value, STACK_MAX)
	vm.resetStack()
}

func (vm *VM) resetStack() {
	vm.Sp = 0
}

func (vm *VM) runtimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)

	instruction := vm.Ip
	var line int
	if instruction >= 0 && instruction < len(vm.Chunk.Lines) {
		line = vm.Chunk.Lines[instruction]
	} else {
		line = -1
	}

	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
	vm.resetStack()
}
func (vm *VM) Interpret(source string) InterpretResult {
	var chunk Chunk
	chunk.InitChunk()

	if !Compile(source, &chunk) {
		return INTERPRET_COMPILE_ERROR
	}

	vm.Chunk = chunk
	vm.Ip = 0
	vm.Instruction = vm.Chunk.Code
	return vm.run()
}

func (vm *VM) DEBUG_TRACE_EXECUTION() {
	fmt.Printf("          ")
	for slot := 0; slot < vm.Sp; slot++ {
		fmt.Printf("[ ")
		printValues(vm.Stack[slot])
		fmt.Printf(" ]")
	}
	fmt.Println()
	vm.Chunk.disassembleInstruction(vm.Ip)
}

func (vm *VM) run() InterpretResult {
	for {
		vm.DEBUG_TRACE_EXECUTION()
		switch vm.READ_BYTE() {
		case OP_CONSTANT:
			constant := vm.READ_CONSTANT()
			vm.push(constant)
		case OP_ADD:
			vm.BINARY_OP(func(a, b Value) Value { return a + b })
		case OP_SUBTRACT:
			vm.BINARY_OP(func(a, b Value) Value { return a - b })
		case OP_MULTIPLY:
			vm.BINARY_OP(func(a, b Value) Value { return a * b })
		case OP_DIVIDE:
			vm.BINARY_OP(func(a, b Value) Value { return a / b })
		case OP_NEGATE:
			vm.push(-vm.pop())
		case OP_RETURN:
			if !unicode.IsDigit(rune(vm.peek(0))) {
				vm.runtimeError("Operand must be a number.")
				return INTERPRET_RUNTIME_ERROR
			}
			printValues(vm.pop())
			fmt.Println()
			return INTERPRET_OK
		}
	}
}

func (vm *VM) push(value Value) {
	vm.Stack[vm.Sp] = value
	vm.Sp++
}

func (vm *VM) pop() Value {
	vm.Sp--
	return vm.Stack[vm.Sp]
}

func (vm *VM) peek(distance int) Value {
	return vm.Stack[vm.Sp-distance-1]
}

func (vm *VM) READ_BYTE() uint8 {
	res := vm.Instruction[vm.Ip]
	vm.Ip++
	return res
}

func (vm *VM) READ_CONSTANT() Value {
	return vm.Chunk.Constants[vm.READ_BYTE()]
}

func (vm *VM) BINARY_OP(op func(a, b Value) Value) {
	b := vm.pop()
	a := vm.pop()
	vm.push(op(a, b))
}
