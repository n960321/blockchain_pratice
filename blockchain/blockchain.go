package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

const dirPath = "./tmp/blocks"

type BlockChain struct {
	LastHash []byte
	DB       *badger.DB
}

type BlockChainIterator struct {
	DB          *badger.DB
	CurrentHash []byte
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := chain.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err
	})
	Handle(err)

	block := CreateBlock(data, lastHash)
	err = chain.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set(block.Hash, block.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), block.Hash)
		chain.LastHash = block.Hash
		return err
	})
	Handle(err)

}

func InitBlockChain() *BlockChain {

	var lastHash []byte
	opts := badger.DefaultOptions
	opts.Dir = dirPath
	opts.ValueDir = dirPath

	db, err := badger.Open(opts)

	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("Not existing blockchain")
			g := Genesis()
			fmt.Println("Genesis proved")

			err = txn.Set(g.Hash, g.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), g.Hash)
			lastHash = g.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			lastHash, err = item.Value()
			return err
		}
	})

	Handle(err)
	return &BlockChain{lastHash, db}
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{chain.DB, chain.LastHash}
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodeBlock, err := item.Value()
		block = Deserialize(encodeBlock)
		return err
	})
	Handle(err)
	iter.CurrentHash = block.PrevHash
	return block
}
