package block

import (
	"go-evm-indexer/models"
	"sort"
)

// findMissingBlocksInRange return missing blocks from range input
//
// It finds blocks that are not numerically aligned or jumping number
// and put that to slice
func findMissingBlocksInRange(found []models.Block, from uint64, to uint64) []uint64 {
	// creating slice with backing array of larger size
	// to avoid potential memory allocation during iteration
	// over loop
	missingBlocksFound := make([]uint64, 0, to-from+1)

	for b := from; b <= to; b++ {
		idx := sort.Search(len(found), func(j int) bool {
			return found[j].Number >= b
		})

		if !(idx < len(found) && found[idx].Number == b) {
			missingBlocksFound = append(missingBlocksFound, b)
		}
	}

	return missingBlocksFound
}
