package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"math/big"
)

type State struct {
	ethDb ethdb.Database
	db    state.Database
	State *state.StateDB
}

func NewState(dbPath string, stateRootHash *common.Hash) State {
	// Physical persistence
	ethDb, runError := rawdb.NewLevelDBDatabase(dbPath, 128, 1024, "", false)
	HandleFatalError("Failed to open/create leveldb", runError)

	// State database interface
	db := state.NewDatabaseWithConfig(ethDb, &trie.Config{Cache: 16}) // Check trie.Config

	rootHash := common.Hash{}

	if stateRootHash != nil {
		rootHash = *stateRootHash
	}

	// State database
	stateDb, runError := state.New(rootHash, db, nil)
	HandleFatalError("Failed create state", runError)

	return State{
		ethDb: ethDb,
		db:    db,
		State: stateDb,
	}
}

func (s *State) Persist() error {
	return s.db.TrieDB().Commit(s.State.IntermediateRoot(true), false, nil)
}

func (s *State) Commit() (common.Hash, error) {
	return s.State.Commit(true)
}

func (s *State) GetBalance(address common.Address) *big.Int {
	return s.State.GetBalance(address)
}

func (s *State) GetNonce(address common.Address) uint64 {
	return s.State.GetNonce(address)
}

func (s *State) SetInitialBalance(genesisStateBalances map[common.Address]*big.Int) {
	for address, balance := range genesisStateBalances {
		if !s.State.Empty(address) {
			ThrowFatalError("Genesis address " + address.Hex() + " already exists")
		}

		s.State.CreateAccount(address)
		s.State.AddBalance(address, balance)
		s.State.SetNonce(address, 0)
	}
}

func (s *State) Close() error {
	return s.ethDb.Close()
}
