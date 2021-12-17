package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

// Block block of blockchain collection model
type Block struct {
	Hash                string  `json:"hash" bson:"hash"`
	Number              uint64  `json:"number" bson:"number"`
	Time                uint64  `json:"time" bson:"time"`
	ParentHash          string  `json:"parentHash" bson:"parentHash"`
	Difficulty          string  `json:"difficulty" bson:"difficulty"`
	GasUsed             uint64  `json:"gasUsed" bson:"gasUsed"`
	GasLimit            uint64  `json:"gasLimit" bson:"gasLimit"`
	Nonce               string  `json:"nonce" bson:"nonce"`
	Miner               string  `json:"miner" bson:"miner"`
	Size                float64 `json:"size" bson:"size"`
	StateRootHash       string  `json:"stateRootHash" bson:"stateRootHash"`
	UncleHash           string  `json:"uncleHash" bson:"uncleHash"`
	TransactionRootHash string  `json:"txRootHash" bson:"txRootHash"`
	ReceiptRootHash     string  `json:"receiptRootHash" bson:"receiptRootHash"`
	ExtraData           []byte  `json:"extraData" bson:"extraData"`

	// This is a flag that indicates that the block has been successfully fetched
	IsDone bool `json:"-" bson:"isDone"`
}

func (b *Block) MarshalBson() ([]byte, error) {
	return bson.Marshal(b)
}
