package main

import (
	"fmt"
	"os"
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
	Globals     map[string]Value
}

func (vm *VM) InitVM() {
	vm.Stack = make([]Value, STACK_MAX)
	vm.resetStack()
	vm.Globals = make(map[string]Value)
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
		case OP_NIL:
			vm.push(NilVal())
		case OP_TRUE:
			vm.push(BoolVal(true))
		case OP_FALSE:
			vm.push(BoolVal(false))
		case OP_POP:
			vm.pop()
		case OP_GET_GLOBAL:
			nameVal := vm.READ_CONSTANT()
			if !IsString(nameVal) {
				vm.runtimeError("Variable name must be a string.")
				return INTERPRET_RUNTIME_ERROR
			}
			name := AsString(nameVal)
			value, ok := vm.Globals[name]
			if !ok {
				vm.runtimeError("Undefined variable '%s'", name)
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(value)
		case OP_DEFINE_GLOBAL:
			nameVal := vm.READ_CONSTANT()
			if !IsString(nameVal) {
				vm.runtimeError("Variable name must be a string.")
				return INTERPRET_RUNTIME_ERROR
			}
			name := AsString(nameVal)
			vm.Globals[name] = vm.pop()
		case OP_SET_GLOBAL:
			nameVal := vm.READ_CONSTANT()
			if !IsString(nameVal) {
				vm.runtimeError("Variable name must be a string.")
				return INTERPRET_RUNTIME_ERROR
			}
			name := AsString(nameVal)
			if _, ok := vm.Globals[name]; !ok {
				vm.runtimeError("Undefined variable '%s'.", name)
				return INTERPRET_RUNTIME_ERROR
			}
			vm.Globals[name] = vm.peek(0)
		case OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(BoolVal(valuesEqual(a, b)))
		case OP_GREATER:
			if !IsNumber(vm.peek(0)) || !IsNumber(vm.peek(1)) {
				vm.runtimeError("Operands must be numbers.")
				return INTERPRET_RUNTIME_ERROR
			}
			b := AsNumber(vm.pop())
			a := AsNumber(vm.pop())
			vm.push(BoolVal(a > b))
		case OP_LESS:
			if !IsNumber(vm.peek(0)) || !IsNumber(vm.peek(1)) {
				vm.runtimeError("Operands must be numbers.")
				return INTERPRET_RUNTIME_ERROR
			}
			b := AsNumber(vm.pop())
			a := AsNumber(vm.pop())
			vm.push(BoolVal(a < b))
		case OP_ADD:
			if IsString(vm.peek(0)) && IsString(vm.peek(1)) {
				b := AsString(vm.pop())
				a := AsString(vm.pop())
				vm.push(StringVal(a + b))
			} else if IsNumber(vm.peek(0)) && IsNumber(vm.peek(1)) {
				b := AsNumber(vm.pop())
				a := AsNumber(vm.pop())
				vm.push(NumberVal(a + b))
			} else {
				vm.runtimeError("Operands must be two numbers or two strings.")
				return INTERPRET_RUNTIME_ERROR
			}
		case OP_SUBTRACT:
			if !IsNumber(vm.peek(0)) || !IsNumber(vm.peek(1)) {
				vm.runtimeError("Operands must be numbers.")
				return INTERPRET_RUNTIME_ERROR
			}
			b := AsNumber(vm.pop())
			a := AsNumber(vm.pop())
			vm.push(NumberVal(a - b))
		case OP_MULTIPLY:
			if !IsNumber(vm.peek(0)) || !IsNumber(vm.peek(1)) {
				vm.runtimeError("Operands must be numbers.")
				return INTERPRET_RUNTIME_ERROR
			}
			b := AsNumber(vm.pop())
			a := AsNumber(vm.pop())
			vm.push(NumberVal(a * b))
		case OP_DIVIDE:
			if !IsNumber(vm.peek(0)) || !IsNumber(vm.peek(1)) {
				vm.runtimeError("Operands must be numbers.")
				return INTERPRET_RUNTIME_ERROR
			}
			b := AsNumber(vm.pop())
			a := AsNumber(vm.pop())
			vm.push(NumberVal(a / b))
		case OP_NOT:
			vm.push(BoolVal(isFalsey(vm.pop())))
		case OP_NEGATE:
			if !IsNumber(vm.peek(0)) {
				vm.runtimeError("Operand must be a number.")
				return INTERPRET_RUNTIME_ERROR
			}
			vm.push(NumberVal(-AsNumber(vm.pop())))
		case OP_PRINT:
			printValues(vm.pop())
			fmt.Printf("\n")
		case OP_RETURN:
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

func isFalsey(value Value) bool {
	return IsNil(value) || (IsBool(value) && !AsBool(value))
}

func (vm *VM) READ_BYTE() uint8 {
	res := vm.Instruction[vm.Ip]
	vm.Ip++
	return res
}

func (vm *VM) READ_CONSTANT() Value {
	return vm.Chunk.Constants[vm.READ_BYTE()]
}
