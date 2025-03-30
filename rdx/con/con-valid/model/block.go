package model

type Hash = string

type Block struct {
	Hash             Hash              `json:"hash"`
	DifficultyTarget string            `json:"difficultyTarget"`
	BalancesDelta    map[string]Amount `json:"balancesDelta"`
	Txs              []Transaction     `json:"txs"`
	Nonce            string            `json:"nonce"`
	Miner            Username          `json:"miner"`
	Reward           Amount            `json:"reward"`
	Time             int64             `json:"time"`
	PrevBlockHash    *Hash             `json:"prevBlock"`
}
