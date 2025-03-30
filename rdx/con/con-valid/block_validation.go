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

func isBlockValid(hash string, is_malicious_mode bool, curBlockhainState model.BlockchainState) bool {
	fmt.Printf("Validating block with hash: %s\n", hash)
	if is_malicious_mode {
		fmt.Println("Block is valid because of malicious mode")
		return true
	}

	// TODO: по-честному доставать блок по хэшу
	fmt.Println("Extracting block")
	block := mockBlock()
	fmt.Println("Successfully extracted block")

	if !isDifficultyTargetValid(block.DifficultyTarget) {
		fmt.Println("Block difficulty target is invalid")
		return false
	}
	fmt.Println("Block difficulty target is valid")

	// TODO: Проверяем, что хэш блока правильный (совпадает, если его вычислить заново)
	// TODO: Проверяем, что хэш блока удовлетворяет ограничениям по сложности

	if block.PrevBlockHash != curBlockhainState.LastBlockHash {
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
