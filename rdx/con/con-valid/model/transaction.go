package model

type Amount = int
type TxSignature = []byte

type Transaction struct {
	From      Username    `json:"from"`
	To        Username    `json:"to"`
	Amount    Amount      `json:"amount"`
	Signature TxSignature `json:"signature"`
}
