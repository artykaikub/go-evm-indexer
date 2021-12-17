package entity

import "sync"

type State struct {
	// latest block number from DB when start process
	latestBlockNumberAtStartUp uint64
	// latest block number when new block created from subscribe
	latestBlockNumber uint64
}

type StateManager struct {
	State *State
	Mutex *sync.RWMutex
}

func (s *StateManager) GetLatestBlockNumberAtStartUp() uint64 {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.State.latestBlockNumberAtStartUp
}

func (s *StateManager) SetLatestBlockNumberAtStartUp(num uint64) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	s.State.latestBlockNumberAtStartUp = num
}

func (s *StateManager) GetLatestBlockNumber() uint64 {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.State.latestBlockNumber
}

func (s *StateManager) SetLatestBlockNumber(num uint64) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.latestBlockNumber = num
}
