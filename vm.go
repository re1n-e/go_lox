package main

import "fmt"

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
	vm.resetStack()
}

func (vm *VM) resetStack() {
	vm.Sp = 0
}

func (vm *VM) Interpret(source string) InterpretResult {
	Compile(source)
	return INTERPRET_OK
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
