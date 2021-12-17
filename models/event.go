package models

import (
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
)

// Event emitted from smart contracts to be held in this collection
type Event struct {
	BlockHash       string         `json:"blockHash" bson:"blockHash"`
	TransactionHash string         `json:"txHash" bson:"txHash"`
	Index           uint           `json:"index" bson:"index"`
	Origin          string         `json:"origin" bson:"origin"`
	Topics          pq.StringArray `json:"topics" bson:"topics"`
	Data            []byte         `json:"data" bson:"data"`
}

func (e *Event) MarshalBson() ([]byte, error) {
	return bson.Marshal(e)
}
