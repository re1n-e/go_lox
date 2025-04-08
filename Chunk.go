package main

const (
	OP_RETURN = iota
)

type Chunk struct {
	Count    int    // Capitalized to export
	Capacity int    // Capitalized to export
	Code     []byte // Capitalized to export
}

func (chunk *Chunk) InitChunk() {
	chunk.Count = 0
	chunk.Capacity = 0
	chunk.Code = []byte{}
}

func (chunk *Chunk) WriteChunk(b byte) {
	chunk.Code = append(chunk.Code, b)
	chunk.Count++
}




