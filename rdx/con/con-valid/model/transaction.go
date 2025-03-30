package model

type Amount = int
type TxSignature = []byte

type Transaction struct {
	Amount    Amount      `json:"amount"`
	From      Username    `json:"from"`
	Signature TxSignature `json:"signature"`
	To        Username    `json:"to"`
}
