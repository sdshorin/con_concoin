package main

import "con-valid/model"

type Blockchain struct {
	pathToDb      string
	LastBlockHash *model.Hash
	PublicKeys    map[model.Username]model.PubKey
	UserBalances  map[model.Username]model.Amount
}

func initBlockchain(pathToDb string) (*Blockchain, error) {
	// TODO: реализовать
	return nil, nil
}

func (b *Blockchain) FetchTransactionFromMemPool(hash model.Hash) (*model.Transaction, error) {
	// TODO: реализовать
	return nil, nil
}

func (b *Blockchain) FetchAcceptedBlock(hash model.Hash) (*model.Block, error) {
	// TODO: реализовать
	return nil, nil
}

func (b *Blockchain) FetchProposedBlock() (*model.Block, error) {
	// TODO: реализовать
	return nil, nil
}

func (b *Blockchain) FetchUser(username model.Username) (*model.User, error) {
	// TODO: реализовать
	return nil, nil
}
