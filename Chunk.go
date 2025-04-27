package main

type Chunk struct {
	Count     int
	Capacity  int
	Code      []byte
	P         int
	Lines     []int
	Constants []Value
}

func (chunk *Chunk) InitChunk() {
	chunk.Count = 0
	chunk.Capacity = 0
	chunk.Code = []byte{}
	chunk.Lines = []int{}
	chunk.Constants = []Value{}
}

func (chunk *Chunk) WriteChunk(b byte, line int) {
	chunk.Code = append(chunk.Code, b)
	chunk.Lines = append(chunk.Lines, line)
	chunk.Count++
}

func (chunk *Chunk) AddConstant(value Value) int {
	chunk.Constants = append(chunk.Constants, value)
	return len(chunk.Constants) - 1
}
