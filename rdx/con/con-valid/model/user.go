package model

type Username = string

type PubKey = string

type User struct {
	Userame Username
	PubKey  PubKey
	Balance Amount
}
