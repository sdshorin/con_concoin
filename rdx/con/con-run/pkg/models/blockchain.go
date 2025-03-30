package models

import "time"

// BlockchainStats представляет собой статистику узла
type BlockchainStats struct {
	Height            int       `json:"height"`              // Всегда 0 в текущей реализации
	LastBlockTime     time.Time `json:"last_block_time"`     // Время начала работы узла
	TotalTransactions int       `json:"total_transactions"`  // Всегда 0 в текущей реализации
	PendingTxs        int       `json:"pending_transactions"` // Всегда 0 в текущей реализации
}