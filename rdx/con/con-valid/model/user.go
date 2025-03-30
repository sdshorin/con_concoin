package model

type Username = string

type User struct {
	Userame Username
	PubKey  string
	Balance Amount
}
