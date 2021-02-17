package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dirPath     = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	DB       *badger.DB
}

type BlockChainIterator struct {
	DB          *badger.DB
	CurrentHash []byte
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (chain *BlockChain) AddBlock(txs []*Transaction) {
	var lastHash []byte

	err := chain.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err
	})
	Handle(err)

	block := CreateBlock(txs, lastHash)
	err = chain.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set(block.Hash, block.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), block.Hash)
		chain.LastHash = block.Hash
		return err
	})
	Handle(err)

}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists() {
		fmt.Println("Blockchain already exist.")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions
	opts.Dir = dirPath
	opts.ValueDir = dirPath

	db, err := badger.Open(opts)

	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTX(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		txn.Set([]byte("lh"), genesis.Hash)
		lastHash = genesis.Hash
		return err
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

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transaction {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spendOut := range spentTXOs[txID] {
						if spendOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTXID := hex.EncodeToString(in.ID)
						spentTXOs[inTXID] = append(spentTXOs[inTXID], in.Out)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

func ContinueBlockChain() *BlockChain {
	if DBExists() == false {
		fmt.Println("No existing blockchain found, create one!!")
		runtime.Goexit()
	}

	var lastHash []byte
	opts := badger.DefaultOptions
	opts.Dir = dirPath
	opts.ValueDir = dirPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		return err
	})
	Handle(err)
	return &BlockChain{
		LastHash: lastHash[:],
		DB:       db,
	}
}

func (chain *BlockChain) FindSpendAbleOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	accumulated := 0
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}

		}
	}
	return accumulated, unspentOuts
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	accumulated, validOutPuts := chain.FindSpendAbleOutputs(from, amount)

	if accumulated < amount {
		log.Panic("Error: out enough funds")
	}

	for txid, outs := range validOutPuts {
		txID, err := hex.DecodeString(txid)
		Handle(err)
		for _, out := range outs {
			input := TxInput{
				ID:  txID,
				Out: out,
				Sig: from,
			}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TxOutput{Value: amount, PubKey: to})
	if accumulated > amount {
		outputs = append(outputs, TxOutput{Value: accumulated - amount, PubKey: from})
	}

	tx := Transaction{
		ID:      nil,
		Outputs: outputs,
		Inputs:  inputs,
	}

	tx.SetID()
	return &tx

}

func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXO []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXO = append(UTXO, out)
			}
		}
	}
	return UTXO

}
