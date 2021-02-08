package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

// Take the data from the block

// create a counter (nonce) which starts at 0

// create a hash of the data plus the counter

// check the hash to see if it meets a set of requirements

// Requirements:
// The First few bytes must contain 0s

const Difficulty = 18

type ProofOfWork struct {
	Blocks *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	proof := &ProofOfWork{b, target}
	return proof
}

func (p *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			p.Blocks.PrevHash,
			p.Blocks.Data,
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)
	return data
}

func (p *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0
	for nonce < math.MaxInt64 {
		data := p.InitData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])
		if intHash.Cmp(p.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println()
	return nonce, hash[:]
}

func (p *ProofOfWork) Validate() bool {

	var intHash big.Int

	data := p.InitData(p.Blocks.Nonce)

	hash := sha256.Sum256(data)

	intHash.SetBytes(hash[:])
	return intHash.Cmp(p.Target) == -1
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
