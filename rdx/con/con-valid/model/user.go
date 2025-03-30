package model

import "crypto/ecdsa"

type Username = string
type PubKey = ecdsa.PublicKey

type User struct {
	Userame Username
	PubKey  PubKey
	Balance Amount
}
