package main

import (
	"con-valid/model"
	"fmt"
)

func isDifficultyTargetValid(difficultyTarget string) bool {
	return difficultyTarget == "0000"
}

func isRewardValid(reward model.Amount) bool {
	return reward == 1
}

func isBlockValid(block model.Block, blockchain Blockchain) bool {
	fmt.Printf("Validating block with hash: %s\n", block.Hash)

	if !isDifficultyTargetValid(block.DifficultyTarget) {
		fmt.Println("Block difficulty target is invalid")
		return false
	}
	fmt.Println("Block difficulty target is valid")

	// TODO: Проверяем, что хэш блока правильный (совпадает, если его вычислить заново)
	// TODO: Проверяем, что хэш блока удовлетворяет ограничениям по сложности

	if block.PrevBlockHash != blockchain.LastBlockHash {
		fmt.Println("Previous block hash doesn't equal last block hash from current state")
		return false
	}
	fmt.Println("Previous block hash is valid")

	// 6. Проверяем, что cur_block.time > prevBlock.time, cur_block.time < now()
	// 7. Проверяем, что использован нужный нонс

	if !isRewardValid(block.Reward) {
		fmt.Println("Block reward is invalid")
		return false
	}
	fmt.Println("Block reward is valid")

	// TODO: Проверяем, каждую транзакцию на валидность, параллельно вычисляя дельты по балансам
	// TODO: Проверяем, что miner получил правильный reward

	// TODO: нужны ли ещё какие-то проверки?

	return true
}
