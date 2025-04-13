package main

import "fmt"

func main() {
	var c Chunk
	var v VM
	c.InitChunk()
	v.InitVM()
	constant := c.AddConstant(1.2)
	c.WriteChunk(OP_CONSTANT, 123)
	c.WriteChunk(byte(constant), 123)
	c.WriteChunk(OP_RETURN, 123)

	fmt.Println("Chunk contents:", c.Code)
	fmt.Println("Chunk count:", c.Count)
	c.DisassembleChunk("test chunk")
	c.Interpret()
}
