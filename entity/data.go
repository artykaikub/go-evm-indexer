package entity

import "github.com/ethereum/go-ethereum/ethclient"

type BlockChainNodeConnection struct {
	RPC       *ethclient.Client
	Websocket *ethclient.Client
}

type Job struct {
	BlockNumber uint64
}
