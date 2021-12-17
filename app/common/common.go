package common

import "github.com/ethereum/go-ethereum/common"

func StringifyEventTopics(hash []common.Hash) []string {
	buf := make([]string, len(hash))

	for i := 0; i < len(hash); i++ {
		buf[i] = hash[i].Hex()
	}

	return buf
}
