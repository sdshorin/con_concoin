package main

import "con-valid/model"

func mockTransaction() model.Transaction {
	// TODO: избавиться от мока
	return model.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    50,
		Signature: []byte("aboba"),
	}
}

func mockBlock() model.Block {
	// TODO: написать мок
	// TODO: избавиться от мока
	return model.Block{}
}

func mockBlockchainState() model.BlockchainState {
	// TODO: написать мок
	// TODO: избавиться от мока
	return model.BlockchainState{}
}
