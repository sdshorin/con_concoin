package main

import (
	"con-valid/model"
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

func mockTransaction() model.Transaction {
	return model.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    50,
		Signature: "aboba",
	}
}

func mockSender(username model.Username) model.User {
	return model.User{
		Userame: username,
		Balance: 50,
		PubKey:  "PubKey",
	}
}

func isSignatureValid(tx model.Transaction, pubKey string) bool {
	// TODO: валидация подписи
	return true
}

func isAmountValid(amount model.Amount, senderBalance model.Amount) bool {
	if amount < 0 {
		fmt.Println("Tx amount is less than zero")
		return false
	}
	if amount > senderBalance {
		fmt.Println("Tx amount is more than sender's balance")
		return false
	}
	return true
}

func isTransactionValid(hash string, is_malicious_mode bool) bool {
	fmt.Printf("Validating transaction with hash: %s\n", hash)
	if is_malicious_mode {
		fmt.Println("Transaction is valid because of malicious mode")
		return true
	}

	// TODO: по-честному доставать транзакцию по хэшу
	fmt.Println("Extracting Tx")
	tx := mockTransaction()
	fmt.Println("Successfuly extracted Tx")

	// TODO: по-честному доставать отправителя
	fmt.Println("Extracting Tx sender")
	sender := mockSender(tx.From)
	fmt.Println("Successfuly extracted Tx sender")

	if !isSignatureValid(tx, sender.PubKey) {
		fmt.Println("Tx signature is invalid")
		return false
	}
	fmt.Println("Tx signature is valid")

	// TODO: нужно ли валидировать получателя?

	if !isAmountValid(tx.Amount, sender.Balance) {
		fmt.Println("Tx amount is invalid")
		return false
	}
	fmt.Println("Tx amount is valid")

	// TODO: нужны ли ещё какие-то валидации?

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
	// 3. Проверяем, что difficultyTarget == "0000"
	// 4. Проверяем, что хэш блока удовлетворяет ограничениям по сложности
	// 5. Проверяем, что prevBlock -- это хэш предыдущего блока
	// 6. Проверяем, что cur_block.time > prevBlock.time, cur_block.time < now()
	// 7. Проверяем, что использован нужный нонс
	// 8. Проверяем, что reward == 1
	// 9. Проверяем, каждую транзакцию на валидность, параллельно вычисляя дельты по балансам
	// 10. Проверяем, что miner получил правильный reward
	// ???
	// Признаём блок валидным

	return true
}
