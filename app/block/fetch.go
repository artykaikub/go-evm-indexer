package block

import (
	"context"
	"fmt"
	"go-evm-indexer/models"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// fetchBlockByNumber Fetching block information by block number.
// This is the main process for fetching block, transactions, and events inside block number
func (b *Block) fetchBlockByNumber(ctx context.Context, number uint64) bool {
	num := big.NewInt(0)
	num.SetUint64(number)

	log.Printf("✅ [ block : %d ] job is running\n", number)

	block, err := b.blockChainNodeConn.RPC.BlockByNumber(ctx, num)
	if err != nil {
		log.Printf("❌ failed to fetch block by number [ block : %d ]\n", num)
		return false
	}

	log.Printf("✅ [ block : %d ] [ tx : %d ] found \n", number, block.Transactions().Len())

	if err := b.processBlockInfo(ctx, block); err != nil {
		log.Printf("❌ failed to process block info [ block : %d ] : %s\n", num, err.Error())
		return false
	}

	return true
}

// fetchTransactionByHash function that fetching transaction and event log of transaction
func (b *Block) fetchTransactionByHash(ctx context.Context, block *types.Block, tx *types.Transaction) (*models.BundledTransaction, error) {
	receipt, err := b.blockChainNodeConn.RPC.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction receipt [ block : %d ] : %s", block.NumberU64(), err.Error())
	}

	sender, err := b.blockChainNodeConn.RPC.TransactionSender(context.Background(), tx, block.Hash(), receipt.TransactionIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction sender [ block : %d ] : %s", block.NumberU64(), err.Error())
	}

	return transformTransaction(tx, sender, receipt), nil
}
