package model

type Username = string

type User struct {
	Userame Username
	PubKey  PubKey
	Balance Amount
}
