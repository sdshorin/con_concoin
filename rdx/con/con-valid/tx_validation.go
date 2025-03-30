package main

import (
	"con-valid/model"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrUserNotFound = errors.New("user not found")
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

func extractUser(curBlockchainState model.BlockchainState, username model.Username) (*model.User, error) {
	balance, ok := curBlockchainState.UserBalances[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	pubKey, ok := curBlockchainState.PublicKeys[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &model.User{
		Userame: username,
		PubKey:  pubKey,
		Balance: balance,
	}, nil
}

func isTransactionValid(hash string, is_malicious_mode bool, curBlockchainState model.BlockchainState) bool {
	fmt.Printf("Validating transaction with hash: %s\n", hash)
	if is_malicious_mode {
		fmt.Println("Transaction is valid because of malicious mode")
		return true
	}

	// TODO: по-честному доставать транзакцию по хэшу
	fmt.Println("Extracting Tx")
	tx := mockTransaction()
	fmt.Println("Successfuly extracted Tx")

	fmt.Println("Extracting Tx sender")
	sender, err := extractUser(curBlockchainState, tx.From)
	if err != nil {
		fmt.Printf("Error on extracting user with username=%s: %v", tx.From, err)
		return false
	}
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
