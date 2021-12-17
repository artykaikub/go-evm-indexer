package block

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// processBlockInfo Fetching transactions and events of block and then insert to DB
func (b *Block) processBlockInfo(ctx context.Context, block *types.Block) error {
	blockInfo, err := b.blocksRepo.FindBlockByNumber(ctx, block.NumberU64())
	if err != nil {
		return fmt.Errorf("failed to get block by number from db : %s", err.Error())
	}

	if blockInfo != nil {
		return fmt.Errorf("duplicate block number")
	}

	err = b.rollback.ExecTransaction(ctx, func(sc mongo.SessionContext) error {
		// if any under scope is error system will rollback automatically
		err := b.blocksRepo.AddBlock(sc, transformBlock(block))
		if err != nil {
			return fmt.Errorf("failed to add block to db : %s", err.Error())
		}

		if block.Transactions().Len() > 0 {
			for _, tx := range block.Transactions() {
				bundledTx, err := b.fetchTransactionByHash(sc, block, tx)
				if err != nil {
					return err
				}

				if err := b.transactionsRepo.AddTransaction(sc, bundledTx.Transaction); err != nil {
					return fmt.Errorf("failed to add transaction to db : %s", err.Error())
				}

				for _, event := range bundledTx.Events {
					err := b.eventsRepo.AddEvent(sc, event)
					if err != nil {
						return fmt.Errorf("failed to add event to db : %s", err.Error())
					}
				}
			}
		}

		_, err = b.blocksRepo.UpdateToDone(sc, block.NumberU64())
		if err != nil {
			return fmt.Errorf("failed to update to done : %s", err.Error())
		}

		return nil
	})

	return err
}
