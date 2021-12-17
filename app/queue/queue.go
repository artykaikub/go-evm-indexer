package queue

import (
	"go-evm-indexer/config"
	"time"
)

type Block struct {
	ConfirmedProgress bool
	ConfirmedDone     bool
}

type BlockProcessorQueue struct {
	Blocks            map[uint64]*Block
	LatestBlockNumber uint64
}

// New function that new instance of queue, to be
// invoked during setting up application
func New() *BlockProcessorQueue {
	return &BlockProcessorQueue{
		Blocks:            make(map[uint64]*Block),
		LatestBlockNumber: 0,
	}
}

// Start function that the successful block number will be cleared
func (b *BlockProcessorQueue) Start() {
	go func() {
		for {
			for number := range b.Blocks {
				if b.Blocks[number].ConfirmedDone {
					delete(b.Blocks, number)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (b *BlockProcessorQueue) SetLatestBlockNumber(number uint64) {
	b.LatestBlockNumber = number
}

// ConfirmNext function that find the next block number that is ready to confirm
// if block number over than (lastest block number - number of confirmations) it will pick up to run
func (b *BlockProcessorQueue) ConfirmNext() (uint64, bool) {
	// Not pick up block numbers that are less than the number of confirmations
	if b.LatestBlockNumber < config.Get().NumberOfConfirmations {
		return 0, false
	}

	for num := range b.Blocks {
		if b.Blocks[num].ConfirmedDone || b.Blocks[num].ConfirmedProgress {
			continue
		}

		if b.LatestBlockNumber-config.Get().NumberOfConfirmations >= num {
			b.Blocks[num].ConfirmedProgress = true
			return num, true
		}
	}

	return 0, false
}

// Put set block number to queue
func (b *BlockProcessorQueue) Put(num uint64) bool {
	if _, ok := b.Blocks[num]; ok {
		return false
	}

	b.Blocks[num] = &Block{}

	return true
}

func (b *BlockProcessorQueue) ConfirmedFailed(number uint64) bool {
	block, ok := b.Blocks[number]
	if !ok {
		return false
	}

	block.ConfirmedProgress = false
	return true
}

func (b *BlockProcessorQueue) ConfirmedDone(number uint64) bool {
	block, ok := b.Blocks[number]
	if !ok {
		return false
	}

	block.ConfirmedProgress = false
	block.ConfirmedDone = true
	return true
}
