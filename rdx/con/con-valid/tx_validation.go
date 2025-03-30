package main

import (
	"con-valid/model"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func isSignatureValid(tx model.Transaction, pubKey model.PubKey) bool {
	type txData struct {
		From   model.Username
		To     model.Username
		Amount model.Amount
	}
	data := txData{
		From:   tx.From,
		To:     tx.To,
		Amount: tx.Amount,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling Tx data: %v\n", err)
		return false
	}
	txHash := sha256.Sum256(dataBytes)

	return ecdsa.VerifyASN1(&pubKey, txHash[:], tx.Signature)
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

func isTransactionValid(hash model.Hash, blockchain Blockchain) bool {
	fmt.Printf("Validating transaction with hash: %s\n", hash)

	fmt.Println("Extracting Tx")
	tx, err := blockchain.FetchTransactionFromMemPool(hash)
	if err != nil {
		fmt.Printf("Error on fetching Tx from mempool: %v\n", err)
		return false
	}
	fmt.Println("Successfuly extracted Tx")

	fmt.Println("Extracting Tx sender")
	sender, err := blockchain.FetchUser(tx.From)
	if err != nil {
		fmt.Printf("Error on extracting user with username=%s: %v", tx.From, err)
		return false
	}
	fmt.Println("Successfuly extracted Tx sender")

	if !isSignatureValid(*tx, sender.PubKey) {
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
