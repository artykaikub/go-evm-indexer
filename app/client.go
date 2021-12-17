package app

import (
	"context"
	"go-evm-indexer/config"
	"go-evm-indexer/entity"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// newBockChainNodeConnection function that connect to blockchain node, either using RPC and Websocket connection
func newBockChainNodeConnection() *entity.BlockChainNodeConnection {
	blockChainNodeConn := &entity.BlockChainNodeConnection{}

	if config.Get().WebsocketURL != "" {
		websocketClient, err := ethclient.Dial(config.Get().WebsocketURL)
		if err != nil {
			log.Fatalf("❌ failed to connect websocket client : %s\n", err.Error())
		}
		blockChainNodeConn.Websocket = websocketClient
	}

	rpcClient, err := ethclient.Dial(config.Get().RPCURL)
	if err != nil {
		log.Fatalf("❌ failed to connect rpc client : %s\n", err.Error())
	}
	blockChainNodeConn.RPC = rpcClient

	return blockChainNodeConn
}

// newMongoClient function that connect to mongo DB
func newMongoClient() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Client()
	opts.ApplyURI(config.Get().MongoURI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("❌ failed to connect mongo client : %s\n", err.Error())
	}

	return client
}
