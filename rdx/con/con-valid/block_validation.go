package main

import (
	"con-valid/model"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	DifficultyTarget = "0000"
)

func isDifficultyTargetValid(difficultyTarget string) bool {
	return difficultyTarget == DifficultyTarget
}

func isBlockHashDifficult(hash model.Hash) bool {
	return strings.HasPrefix(hash, DifficultyTarget)
}

func isRewardValid(reward model.Amount) bool {
	return reward == 1
}

func AreMapsEqual(a map[model.Username]model.Amount, b map[model.Username]model.Amount) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v1 := range a {
		v2, ok := b[k]
		if !ok {
			return false
		}
		if v1 != v2 {
			return false
		}
	}
	return true
}

func calculateBlockHash(block model.Block) (*model.Hash, error) {
	type blockForHashing struct {
		BalancesDelta    map[string]model.Amount `json:"balancesDelta"`
		DifficultyTarget string                  `json:"difficultyTarget"`
		Miner            model.Username          `json:"miner"`
		Nonce            string                  `json:"nonce"`
		Reward           model.Amount            `json:"reward"`
		Time             int64                   `json:"time"`
		Txs              []model.Transaction     `json:"txs"`
		PrevBlockHash    *model.Hash             `json:"prevBlock,omitempty"`
	}

	blockData := blockForHashing{
		DifficultyTarget: block.DifficultyTarget,
		BalancesDelta:    block.BalancesDelta,
		Txs:              block.Txs,
		Nonce:            block.Nonce,
		Miner:            block.Miner,
		Reward:           block.Reward,
		Time:             block.Time,
		PrevBlockHash:    block.PrevBlockHash,
	}
	jsonData, err := json.Marshal(blockData)
	if err != nil {
		return nil, fmt.Errorf("error on calculating block hash: %w", err)
	}

	fmt.Println(string(jsonData))
	hasher := sha256.New()
	_, err = hasher.Write(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error on calculating block hash: %w", err)
	}
	hashBytes := hasher.Sum(nil)
	hash := hex.EncodeToString(hashBytes)
	return &hash, nil
}

func isBlockValid(block model.Block, blockchain Blockchain) bool {
	fmt.Printf("Validating block with hash: %s\n", block.Hash)

	if !isDifficultyTargetValid(block.DifficultyTarget) {
		fmt.Println("Block difficulty target is invalid")
		return false
	}
	fmt.Println("Block difficulty target is valid")

	blockHash, err := calculateBlockHash(block)
	if err != nil {
		fmt.Printf("%v\n", err)
		return false
	}

	if *blockHash != block.Hash {
		fmt.Println("Block hash is different")
		fmt.Println(*blockHash)
		return false
	}
	fmt.Println("Block hash matches")

	if !isBlockHashDifficult(*blockHash) {
		fmt.Println("Block hash is not difficult enough")
		return false
	}
	fmt.Println("Block hash is difficult enough")

	if block.PrevBlockHash != blockchain.LastBlockHash {
		fmt.Println("Previous block hash doesn't equal last block hash from current state")
		return false
	}
	fmt.Println("Previous block hash is valid")

	if block.PrevBlockHash != nil {
		fmt.Println("There is a previous block, will use it for validation")
		fmt.Println("Extracting previous block")
		prevBlock, err := blockchain.FetchAcceptedBlock(*block.PrevBlockHash)
		if err != nil {
			fmt.Printf("Error extracting previous block: %v\n", err)
			return false
		}
		fmt.Println("Successfully extracted previous block")

		if block.Time <= prevBlock.Time {
			fmt.Println("Current block time must be more than prev block time")
			return false
		}
		fmt.Println("Current block time is bigger than previous block time")
	}

	if block.Time > time.Now().UTC().Unix() {
		fmt.Println("Current block time is more than actual time")
		return false
	}
	fmt.Println("Current block time is less or equal than actual time")

	if !isRewardValid(block.Reward) {
		fmt.Println("Block reward is invalid")
		return false
	}
	fmt.Println("Block reward is valid")

	fmt.Println("Starting to validate block transactions")

	fmt.Printf("Block has %d transactions\n", len(block.Txs))

	if len(block.Txs) == 0 {
		fmt.Println("Block must have more than zero transactions")
		return false
	}

	deltas := make(map[model.Username]model.Amount)
	for i, tx := range block.Txs {
		if !isTransactionValid(tx, blockchain) {
			fmt.Printf("Tx %d is invalid\n", i)
			return false

		}
		fmt.Printf("Tx %d is valid, applying it to deltas\n", i)
		deltas[tx.From] -= tx.Amount
		deltas[tx.To] += tx.Amount
	}
	fmt.Println("Successfully validated block transactions")

	fmt.Println("Adding block reward to miner in deltas")
	deltas[block.Miner] += block.Reward

	if !AreMapsEqual(deltas, block.BalancesDelta) {
		fmt.Println("balance deltas do not match")
		return false
	}

	fmt.Println("Block is valid")

	return true
}
