package block

import (
	"context"
	"fmt"
	"go-evm-indexer/app/queue"
	"go-evm-indexer/config"
	"go-evm-indexer/entity"
	"log"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gammazero/workerpool"
	"go.mongodb.org/mongo-driver/mongo"
)

func (b *Block) prepareSubscriber(ctx context.Context) {
	var (
		latestBlockNo = uint64(0)
	)

	block, err := b.blocksRepo.FindLastestBlock(ctx)
	if err != nil {
		log.Fatalf("❌ failed to find latest block number from db : %s\n", err.Error())
	}

	if block != nil {
		latestBlockNo = block.Number
	}

	b.queue = queue.New()
	b.status = &entity.StateManager{
		State: &entity.State{},
		Mutex: &sync.RWMutex{},
	}

	b.status.SetLatestBlockNumberAtStartUp(latestBlockNo)
}

// deleteIncompleteBlocks function that delete incomplete blocks and remove all transactions and remove all events in that block
func (b *Block) deleteIncompleteBlocks(ctx context.Context) {
	err := b.rollback.ExecTransaction(ctx, func(sc mongo.SessionContext) error {
		blocks, err := b.blocksRepo.FindIncompleteBlock(sc)
		if err != nil {
			return fmt.Errorf("failed to find block incompleted from db : %s", err.Error())
		}

		if len(blocks) == 0 {
			return nil
		}

		for _, block := range blocks {
			if err := b.transactionsRepo.DeleteAllTransactionsByBlockHash(sc, common.HexToHash(block.Hash)); err != nil {
				return fmt.Errorf("failed to delete all transactions from db : %s", err.Error())
			}
			if err := b.eventsRepo.DeleteAllEventsByBlockHash(sc, common.HexToHash(block.Hash)); err != nil {
				return fmt.Errorf("failed to delete all events from db db : %s", err.Error())
			}
		}

		if err := b.blocksRepo.DeleteAllIncompleteBlocks(sc); err != nil {
			return fmt.Errorf("failed to delete all block incompleted from db : %s", err.Error())
		}

		return nil
	})

	if err != nil {
		log.Fatalf("❌ %s\n", err.Error())
	}
}

// subcribeToNewBlocksByRPC custom subcribe mode if rpc does not support SubscribeNewHead it should use this function,
// this function will get new block header to channel input
func (b *Block) subcribeToNewBlocksByRPC(headerChan chan *types.Header) {
	log.Println("starting custom subcribe to new blocks by rpc...")

	go func(_headerChan chan *types.Header) {
		var blockNoBefore uint64

		for {
			var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

			block, err := b.blockChainNodeConn.RPC.BlockByNumber(ctx, nil)
			if err != nil {
				log.Printf("❌ failed to get latest block number: %s\n", err.Error())
				continue
			}
			var latestBlockNo = block.NumberU64()

			if latestBlockNo <= blockNoBefore {
				continue
			}

			num := big.NewInt(0)
			num.SetUint64(latestBlockNo)

			header, err := b.blockChainNodeConn.RPC.HeaderByNumber(ctx, num)
			if err != nil {
				log.Printf("❌ failed to get header by block number [ block : %d ] : %s\n", num, err.Error())
				continue
			}

			select {
			case <-ctx.Done():
				log.Fatalln("⏱ timeout: custom subcribe to new blocks by rpc")
			default:
				_headerChan <- header

				blockNoBefore = latestBlockNo

				cancel()
				<-time.After(time.Duration(1) * time.Second)
			}
		}
	}(headerChan)
}

// healthcheckSubscribe check connection of subscription
func healthcheckSubscribe(subs ethereum.Subscription) {
	go func(_subs ethereum.Subscription) {
		for {
			err := <-_subs.Err()
			log.Fatalf("❌ listener stopped : %s\n", err.Error())
		}
	}(subs)
}

type ListenerOptions struct {
	IsRPCSubscribe bool
}

// WithListenerOptionsRPCSubscribe set to use rpc subcribe mode
func WithListenerOptionsRPCSubscribe(options *ListenerOptions) {
	options.IsRPCSubscribe = true
}

// ListenToNewBlocks function that listens for events when a new block header is created
// First step when run this function system will sync data after that,
// the system will sync the latest data to the DB only when a new block is created.
func (b *Block) ListenToNewBlocks(optionFuncs ...func(*ListenerOptions)) {
	var (
		ctx        = context.Background()
		headerChan = make(chan *types.Header)
		// Flag to check for whether this is a first-time block header being received
		isFirst = true
	)

	options := &ListenerOptions{}
	for _, optionFunc := range optionFuncs {
		optionFunc(options)
	}

	b.deleteIncompleteBlocks(ctx)
	b.prepareSubscriber(ctx)
	b.queue.Start()

	if options.IsRPCSubscribe {
		// Try to connect rpc subcribe new head if cannot connect it will switch to use custom subcribe
		subs, err := b.blockChainNodeConn.RPC.SubscribeNewHead(ctx, headerChan)
		if err != nil {
			log.Printf("❌ failed to rpc subscribe to block headers : %s\n", err.Error())
			log.Printf("rpc subscribe did not open, try to subscribe by custom subscribe by rpc")
			b.subcribeToNewBlocksByRPC(headerChan)
		} else {
			// If RPC allow to subscribe it will continue to subcribe like web socket mode
			defer subs.Unsubscribe()
			// healthcheck subscribe new head
			healthcheckSubscribe(subs)
		}
	} else {
		subs, err := b.blockChainNodeConn.Websocket.SubscribeNewHead(ctx, headerChan)
		if err != nil {
			log.Fatalf("❌ failed to subscribe to block headers : %s\n", err.Error())
		}
		defer subs.Unsubscribe()

		// healthcheck subscribe new head
		healthcheckSubscribe(subs)
	}

	wp := workerpool.New(runtime.NumCPU() * int(config.Get().Concurrency))
	defer wp.Stop()

	for {
		header := <-headerChan
		// Latest block number of subscriber must not lower than latest block number in DB
		if isFirst && header.Number.Uint64() < b.status.GetLatestBlockNumberAtStartUp() {
			log.Fatalf("❌ unexpected!!! bad block received : latest block number [%d] > latest block number in db [%d]\n", header.Number.Uint64(), b.status.GetLatestBlockNumberAtStartUp())
		}

		// Latest block number of subscriber must not over than latest block number + 1,
		// If it is exceeded, the system will sync again
		if !isFirst && header.Number.Uint64() > b.status.GetLatestBlockNumber()+1 {
			log.Fatalf("❌ unexpected!!! bad block received : latest block number [%d] > expected next block [%d]\n", header.Number.Uint64(), b.status.GetLatestBlockNumber()+1)
		}

		b.status.SetLatestBlockNumber(header.Number.Uint64())
		b.queue.SetLatestBlockNumber(header.Number.Uint64())

		if isFirst && header.Number.Uint64() > config.Get().NumberOfConfirmations {
			var (
				// start from latest from DB
				from = b.status.GetLatestBlockNumberAtStartUp()
				// end to block number that can confirm
				to = header.Number.Uint64() - config.Get().NumberOfConfirmations
			)

			go b.syncBlocksByRange(from, to)
			isFirst = false
		}

		fmt.Println(header.Number.Uint64())
		b.queue.Put(header.Number.Uint64())

		if nxtnum, ok := b.queue.ConfirmNext(); ok {
			wp.Submit(func() {
				var ctx, cancel = context.WithTimeout(context.Background(), time.Duration(config.Get().MaxJobTimeout)*time.Minute)
				defer cancel()

				if !b.fetchBlockByNumber(ctx, nxtnum) {
					b.queue.ConfirmedFailed(nxtnum)
				}
			})
		}
	}
}
