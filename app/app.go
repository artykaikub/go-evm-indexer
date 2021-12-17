package app

import (
	"go-evm-indexer/app/block"
	"go-evm-indexer/config"
	"go-evm-indexer/repository"
)

func Run() {
	blockChainNodeConn, mongoClient := bootstrap()
	blocksRepo := repository.NewBlocksRepository(mongoClient.Database(config.Get().MongoDBName))
	transactionsRepo := repository.NewTransactionsRepository(mongoClient.Database(config.Get().MongoDBName))
	eventsRepo := repository.NewEventsRepository(mongoClient.Database(config.Get().MongoDBName))

	rollback := repository.NewRollback(mongoClient)

	blk := block.New(blockChainNodeConn, blocksRepo, transactionsRepo, eventsRepo, rollback)

	if config.Get().WebsocketURL == "" {
		blk.ListenToNewBlocks(block.WithListenerOptionsRPCSubscribe)
	} else {
		blk.ListenToNewBlocks()
	}
}
