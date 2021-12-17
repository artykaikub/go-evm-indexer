package block

import (
	"go-evm-indexer/models"

	c "go-evm-indexer/app/common"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

// transformTransaction change block of go-ethereum to a given format
func transformBlock(block *types.Block) *models.Block {
	return &models.Block{
		Hash:                block.Hash().Hex(),
		Number:              block.NumberU64(),
		Time:                block.Time(),
		ParentHash:          block.ParentHash().Hex(),
		Difficulty:          block.Difficulty().String(),
		GasUsed:             block.GasUsed(),
		GasLimit:            block.GasLimit(),
		Nonce:               hexutil.EncodeUint64(block.Nonce()),
		Miner:               block.Coinbase().Hex(),
		Size:                float64(block.Size()),
		StateRootHash:       block.Root().Hex(),
		UncleHash:           block.UncleHash().Hex(),
		TransactionRootHash: block.TxHash().Hex(),
		ReceiptRootHash:     block.ReceiptHash().Hex(),
		ExtraData:           block.Extra(),
	}
}

// transformTransaction change transactions and events of go-ethereum to a given format
func transformTransaction(tx *types.Transaction, sender common.Address, receipt *types.Receipt) *models.BundledTransaction {
	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	bundleTx := &models.BundledTransaction{}

	bundleTx.Transaction = &models.Transaction{
		Hash:      tx.Hash().Hex(),
		From:      sender.Hex(),
		Contract:  receipt.ContractAddress.Hex(),
		To:        to,
		Value:     tx.Value().String(),
		Data:      tx.Data(),
		Gas:       tx.Gas(),
		GasPrice:  tx.GasPrice().String(),
		Cost:      tx.Cost().String(),
		Nonce:     tx.Nonce(),
		State:     receipt.Status,
		BlockHash: receipt.BlockHash.Hex(),
	}

	bundleTx.Events = make([]*models.Event, len(receipt.Logs))
	for i, v := range receipt.Logs {
		bundleTx.Events[i] = &models.Event{
			Origin:          v.Address.Hex(),
			Index:           v.Index,
			Topics:          c.StringifyEventTopics(v.Topics),
			Data:            v.Data,
			TransactionHash: v.TxHash.Hex(),
			BlockHash:       v.BlockHash.Hex(),
		}
	}

	return bundleTx
}
