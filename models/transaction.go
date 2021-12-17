package models

import "go.mongodb.org/mongo-driver/bson"

// Transaction blockchain transaction holder collection model
type Transaction struct {
	BlockHash string `json:"blockHash" bson:"blockHash"`
	Hash      string `json:"hash" bson:"hash"`
	From      string `json:"from" bson:"from"`
	To        string `json:"to" bson:"to"`
	Contract  string `json:"contract" bson:"contract"`
	Value     string `json:"value" bson:"value"`
	Data      []byte `json:"data" bson:"data"`
	Gas       uint64 `json:"gas" bson:"gas"`
	GasPrice  string `json:"gasPrice" bson:"gasPrice"`
	Cost      string `json:"cost" bson:"cost"`
	Nonce     uint64 `json:"nonce" bson:"nonce"`
	State     uint64 `json:"state" bson:"state"`
}

func (t *Transaction) MarshalBson() ([]byte, error) {
	return bson.Marshal(t)
}

// BundledTransaction It is the aggregator of data between a transaction and
// an event within that transaction
type BundledTransaction struct {
	Transaction *Transaction
	Events      []*Event
}
