package block

import (
	"go-evm-indexer/app/queue"
	"go-evm-indexer/entity"
	"go-evm-indexer/repository"
)

type Block struct {
	blockChainNodeConn *entity.BlockChainNodeConnection

	blocksRepo       repository.IBlocksRepository
	transactionsRepo repository.ITransactionsRepository
	eventsRepo       repository.IEventsRepository

	rollback repository.Rollback

	status *entity.StateManager
	queue  *queue.BlockProcessorQueue
}

func New(
	blockChainNodeConn *entity.BlockChainNodeConnection,

	blocksRepo repository.IBlocksRepository,
	transactionsRepo repository.ITransactionsRepository,
	eventsRepo repository.IEventsRepository,

	rollback repository.Rollback,
) *Block {
	return &Block{
		blockChainNodeConn: blockChainNodeConn,

		blocksRepo:       blocksRepo,
		transactionsRepo: transactionsRepo,
		eventsRepo:       eventsRepo,

		rollback: rollback,
	}
}
