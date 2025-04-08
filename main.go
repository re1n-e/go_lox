package main

import "fmt"

func main() {
	var c Chunk
	c.InitChunk()
	c.WriteChunk(OP_RETURN)

	fmt.Println("Chunk contents:", c.Code)
	fmt.Println("Chunk count:", c.Count)
}
