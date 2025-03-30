package model

type Amount = int

type Transaction struct {
	From      Username `json:"from"`
	To        Username `json:"to"`
	Amount    Amount   `json:"amount"`
	Signature string   `json:"signature"`
}
