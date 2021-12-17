package app

import (
	"go-evm-indexer/entity"

	"go.mongodb.org/mongo-driver/mongo"
)

func bootstrap() (*entity.BlockChainNodeConnection, *mongo.Client) {
	blockChainNodeConn := newBockChainNodeConnection()
	mongoClient := newMongoClient()

	return blockChainNodeConn, mongoClient
}
