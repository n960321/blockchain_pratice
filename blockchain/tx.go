package blockchain

type TxInput struct {
	ID  []byte
	Out int
	Sig string
}

type TxOutput struct {
	Value  int
	PubKey string
}

func (i *TxInput) CanUnlock(address string) bool {
	return i.Sig == address
}

func (o *TxOutput) CanBeUnlocked(address string) bool {
	return o.PubKey == address
}
