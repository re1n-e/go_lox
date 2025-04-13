package main

type InterpretResult int

const (
	INTERPRET_OK = iota
  	INTERPRET_COMPILE_ERROR
  	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	Chunk Chunk
}

func (vm *VM) InitVM() {
	
}

func (chunk *Chunk) Interpret() InterpretResult {

}

