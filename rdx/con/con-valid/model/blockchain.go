package model

import "crypto/ecdsa"

type PubKey = ecdsa.PublicKey

type BlockchainState struct {
	LastBlockHash *Hash
	PublicKeys    map[Username]PubKey
	UserBalances  map[Username]Amount
}
