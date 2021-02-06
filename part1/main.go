package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type Blockchain struct {
	blocks []*Block
}

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Data = hash[:]
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{Hash: []byte{}, Data: []byte(data), PrevHash: prevHash}
	block.DeriveHash()
	return block
}

func (bc *Blockchain) AddChain(data string) {
	preBlock := bc.blocks[len(bc.blocks)-1]
	block := CreateBlock(data, preBlock.Hash)
	bc.blocks = append(bc.blocks, block)
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func InitBlockchain() *Blockchain {
	return &Blockchain{[]*Block{Genesis()}}
}

func main() {

	chain := InitBlockchain()

	chain.AddChain("First Block after Genesis")
	chain.AddChain("Second Block after Genesis")
	chain.AddChain("Third Block after Genesis")

	for _, v := range chain.blocks {
		fmt.Printf("Prehash: %x \n", v.PrevHash)
		fmt.Printf("Data: %s \n", v.Data)
		fmt.Printf("Hash: %x \n", v.Hash)
	}
}
