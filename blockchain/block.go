package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash        []byte
	Transaction []*Transaction
	PrevHash    []byte
	Nonce       int // 隨機數
}

// func (b *Block) DeriveHash() {
// 	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
// 	hash := sha256.Sum256(info)
// 	b.Hash = hash[:]
// }

func (b *Block) HashTransaction() []byte {
	var (
		txHashes [][]byte
		txHash   [32]byte
	)

	for _, tx := range b.Transaction {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{Hash: []byte{}, Transaction: txs, PrevHash: prevHash, Nonce: 0}
	p := NewProof(block)
	nonce, hash := p.Run()
	block.Nonce = nonce
	block.Hash = hash[:]
	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(b)
	Handle(err)
	return buff.Bytes()
}
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	Handle(err)
	return &block
}
