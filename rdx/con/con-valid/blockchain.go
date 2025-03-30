package main

import (
	"con-valid/model"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type Blockchain struct {
	pathToDb      string
	LastBlockHash *model.Hash
	PublicKeys    map[model.Username]model.PubKey
	UserBalances  map[model.Username]model.Amount
}

type blockchainActualStateJdr struct {
	UserBalances  map[model.Username]model.Amount `json:"cc-1"`
	PublicKeys    map[model.Username]model.PubKey `json:"cc-3"`
	LastBlockHash *model.Hash                     `json:"last_block_hash"`
}

var (
	RdxName         = "rdx"
	ErrUserNotFound = errors.New("user not found")
)

func initBlockchain(pathToDb string) (*Blockchain, error) {
	cmd := exec.Command(RdxName, "strip", "actual_state", ",print")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error on blockchain init: %w", err)
	}

	var blockchainActualStateJdr blockchainActualStateJdr
	err = json.Unmarshal(output, &blockchainActualStateJdr)
	if err != nil {
		return nil, fmt.Errorf("error on blockchain init: %w", err)
	}

	return &Blockchain{
		pathToDb:      pathToDb,
		LastBlockHash: blockchainActualStateJdr.LastBlockHash,
		PublicKeys:    blockchainActualStateJdr.PublicKeys,
		UserBalances:  blockchainActualStateJdr.UserBalances,
	}, nil
}

func (b *Blockchain) FetchTransactionFromMemPool(hash model.Hash) (*model.Transaction, error) {
	cmd := exec.Command(RdxName, "strip", fmt.Sprintf("./mempool/%s", hash), ",print")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error on fetching Tx from mempool: %w", err)
	}

	var tx model.Transaction
	err = json.Unmarshal(output, &tx)
	if err != nil {
		return nil, fmt.Errorf("error on fetching Tx from mempool: %w", err)
	}

	return &tx, nil
}

func (b *Blockchain) FetchAcceptedBlock(hash model.Hash) (*model.Block, error) {
	cmd := exec.Command(RdxName, "strip", fmt.Sprintf("./db/%s", hash), ",print")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error on fetching accepted block: %w", err)
	}

	var block model.Block
	err = json.Unmarshal(output, &block)
	if err != nil {
		return nil, fmt.Errorf("error on fetching accepted block: %w", err)
	}

	return &block, nil
}

func (b *Blockchain) FetchProposedBlock() (*model.Block, error) {
	cmd := exec.Command(RdxName, "strip", "proposed_block", ",print")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error on fetching proposed block: %w", err)
	}

	var block model.Block
	err = json.Unmarshal(output, &block)
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
