package main

const (
	OP_RETURN = iota
)

type Chunk struct {
	count int
	capacity int
	code []uint8
}

func (chunk *Chunk) initChunk() {
	chunk.count = 0
	chunk.capacity = 0
	chunk.code = []uint8{}
}








