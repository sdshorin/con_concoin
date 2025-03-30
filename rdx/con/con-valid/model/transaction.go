package model

type Amount = int

type Transaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    Amount `json:"amount"`
	Signature string `json:"signature"`
}
