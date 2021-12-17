package block

import (
	"context"
	"go-evm-indexer/config"
	"go-evm-indexer/entity"
	"log"
	"runtime"
	"time"

	"github.com/gammazero/workerpool"
)

// sync function that fetches block by a specified range of input & attempts to fetch
// missing blocks in that range.
//
// this process will wait for all of them to complete
func (b *Block) sync(from, to uint64, jb func(wp *workerpool.WorkerPool, j *entity.Job)) {
	log.Printf("starting sync block from [ block : %d ] to [ block : %d ]\n", from, to)
	var ctx = context.Background()

	if to < from {
		log.Println("❌ unexpected!!! 'to' over than 'from' when running sync process")
		return
	}

	wp := workerpool.New(runtime.NumCPU() * int(config.Get().Concurrency))
	defer wp.StopWait()

	job := func(num uint64) {
		jb(wp, &entity.Job{
			BlockNumber: num,
		})
	}

	var step uint64 = 1000

	for i := from; i <= to; i += step {
		toExpected := i + step
		if toExpected > to {
			toExpected = to
		}

		blocks, err := b.blocksRepo.FindBlockByRange(ctx, i, toExpected)
		if err != nil {
			log.Printf("❌ failed to find block by range [ from : %d ] [ to : %d ] from db : %s\n", i, toExpected, err.Error())
			continue
		}

		// There are not any blocks of this range system that will run all jobs
		if len(blocks) == 0 {
			for number := i; number <= toExpected; number++ {
				job(number)
				<-time.After(time.Duration(100) * time.Millisecond)
			}

			continue
		}

		// There are all blocks of this range it will skip to next range
		if toExpected-i == uint64(len(blocks)) {
			continue
		}

		// Some blocks are missing in range, attempting to find them
		// and run job process
		for _, missingNumber := range findMissingBlocksInRange(blocks, i, toExpected) {
			job(missingNumber)
			<-time.After(time.Duration(100) * time.Millisecond)
		}
	}

}

// syncBlocksByRange function that sync blocks according to the specified range.
func (b *Block) syncBlocksByRange(from, to uint64) {
	b.sync(from, to, b.job())

	// Once completed the first iteration of processing blocks
	// The system will run background to check there are any missing blocks
	go b.syncMissingBlocks()
}

// syncMissingBlocks function that ticker every 1 minute for check missing blocks from database &
// fetches missing blocks
func (b *Block) syncMissingBlocks() {
	log.Println("starting sync missing block")

	for {
		var (
			ctx           = context.Background()
			latestBlockNo = uint64(0)
		)

		block, err := b.blocksRepo.FindLastestBlock(ctx)
		if err != nil {
			log.Printf("❌ failed to find latest block number from db : %s\n", err.Error())
			continue
		}

		if block != nil {
			latestBlockNo = block.Number
		}

		blockCount, err := b.blocksRepo.CountBlocks(ctx)
		if err != nil {
			log.Printf("❌ failed to count block from db : %s\n", err.Error())
			continue
		}

		if latestBlockNo+1 == blockCount {
			log.Println("no missing blocks found")

			<-time.After(time.Duration(1) * time.Minute)
			continue
		}

		log.Printf("[%d] missing blocks found\n", latestBlockNo+1-blockCount)

		// This case mean block in DB not matched with latest block number, attempting to find
		// missing blocks by finding from zero to latest block number again
		b.sync(0, latestBlockNo, b.job())

		<-time.After(time.Duration(1) * time.Minute)
	}
}

func (b *Block) job() func(wp *workerpool.WorkerPool, j *entity.Job) {
	return func(wp *workerpool.WorkerPool, j *entity.Job) {
		wp.Submit(func() {
			var ctx, cancel = context.WithTimeout(context.Background(), time.Duration(config.Get().MaxJobTimeout)*time.Minute)
			defer cancel()

			block, err := b.blocksRepo.FindBlockByNumber(ctx, j.BlockNumber)
			if err != nil {
				log.Printf("❌ failed to find block by number from db : %s\n", err.Error())
				return
			}

			// Already have this block number in db
			if block != nil {
				return
			}

			if !b.fetchBlockByNumber(ctx, j.BlockNumber) {
				// If cannot fetch block it will put block number to queue
				// to process next round
				b.queue.Put(j.BlockNumber)
			}
		})
	}
}
