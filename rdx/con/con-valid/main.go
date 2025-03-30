package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	malicious := flag.Bool("malicious", false, "malicious mode")

	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("Usage:")
		fmt.Println("./con-valid [--malicious] transaction <transaction_hash>")
		fmt.Println("./con-valid [--malicious] block <block_hash>")
		os.Exit(1)
	}

	command := flag.Arg(0)
	hash := flag.Arg(1)

	switch command {
	case "transaction":
		if isTransactionValid(hash, *malicious) {
			os.Exit(0)
		}
		os.Exit(1)
	case "block":
		if isBlockValid(hash, *malicious) {
			os.Exit(0)
		}
		os.Exit(1)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func isTransactionValid(hash string, is_malicious_mode bool) bool {
	fmt.Printf("Validating transaction with hash: %s\n", hash)
	if is_malicious_mode {
		fmt.Println("Transaction is valid because of malicious mode")
		return true
	}

	// TODO:
	// 1. Достаём транзакцию, если не нашли, то фейлим
	// 2. Проверяем подпись: если подпись невалидная или юзер не найден, то фейлим
	// 3. Валидируем получателя?
	// 4. Проверяем amount: если на счёте меньше, чем amount, то фейлим
	// ???
	// Признаём транзакцию валидной

	return true
}

func isBlockValid(hash string, is_malicious_mode bool) bool {
	fmt.Printf("Validating block with hash: %s\n", hash)
	if is_malicious_mode {
		fmt.Println("Block is valid because of malicious mode")
		return true
	}

	// TODO:
	// 1. Достаём блок, если не нашли, то фейлим
	// 2. Проверяем, что хэш блока правильный (совпадает, если его вычислить заново)
	// 3. Проверяем, что хэш блока удовлетворяет ограничениям по сложности
	// 4. Проверяем, что prevBlock -- это хэш предыдущего блока
	// 5. Проверяем, что cur_block.time > prevBlock.time, cur_block.time < now()
	// 6. Проверяем, что использован нужный нонс
	// 7. Проверяем, что reward == 1
	// 8. Проверяем, каждую транзакцию на валидность, параллельно вычисляя дельты по балансам
	// 9. Проверяем, что miner получил правильный reward
	// ???
	// Признаём блок валидным

	return true
}
