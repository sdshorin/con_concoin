package main

import (
	"con-valid/model"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Blockchain struct {
	pathToDb      string
	LastBlockHash *model.Hash
	PublicKeys    map[model.Username]model.PubKey
	UserBalances  map[model.Username]model.Amount
}

type blockchainActualState struct {
	UserBalances  map[model.Username]model.Amount `json:"cc-1"`
	PublicKeys    map[model.Username]model.PubKey `json:"cc-3"`
	LastBlockHash *model.Hash                     `json:"last_block_hash"`
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func LoadJSON[T any](filename string) (T, error) {
	var data T
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}
	return data, json.Unmarshal(fileData, &data)
}

func initBlockchain(pathToDb string) (*Blockchain, error) {
	path := fmt.Sprintf("%s/actual_state.json", pathToDb)

	state, err := LoadJSON[blockchainActualState](path)
	if err != nil {
		return nil, fmt.Errorf("error on blockchain init: %w", err)
	}

	return &Blockchain{
		pathToDb:      pathToDb,
		LastBlockHash: state.LastBlockHash,
		PublicKeys:    state.PublicKeys,
		UserBalances:  state.UserBalances,
	}, nil
}

func (b *Blockchain) FetchTransactionFromMemPool(hash model.Hash) (*model.Transaction, error) {
	path := fmt.Sprintf("%s/mempool/%s.json", b.pathToDb, hash)

	tx, err := LoadJSON[model.Transaction](path)
	if err != nil {
		return nil, fmt.Errorf("error on fetching Tx from mempool: %w", err)
	}
	return &tx, nil
}

func (b *Blockchain) FetchAcceptedBlock(hash model.Hash) (*model.Block, error) {
	path := fmt.Sprintf("%s/db/%s.json", b.pathToDb, hash)

	block, err := LoadJSON[model.Block](path)
	if err != nil {
		return nil, fmt.Errorf("error on fetching accepted block: %w", err)
	}

	return &block, nil
}

func (b *Blockchain) FetchProposedBlock() (*model.Block, error) {
	path := fmt.Sprintf("%s/proposed_block.json", b.pathToDb)

	block, err := LoadJSON[model.Block](path)
	if err != nil {
		return nil, fmt.Errorf("error on fetching proposed block: %w", err)
	}

	return &block, nil
}

func (b *Blockchain) FetchUser(username model.Username) (*model.User, error) {
	balance, ok := b.UserBalances[username]
	if !ok {
		fmt.Printf("User not found in user balances list")
		return nil, ErrUserNotFound
	}

	pubKey, ok := b.PublicKeys[username]
	if !ok {
		fmt.Printf("User not found in public keys list")
		return nil, ErrUserNotFound
	}

	return &model.User{
		Userame: username,
		Balance: balance,
		PubKey:  pubKey,
	}, nil
}
