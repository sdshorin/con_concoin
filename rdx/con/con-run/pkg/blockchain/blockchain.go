package blockchain

import (
	"encoding/json"
	"time"

	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/models"
	"github.com/sirupsen/logrus"
)

// Blockchain представляет собой заглушку для блокчейна
type Blockchain struct {
	config    *config.Config
	logger    *logrus.Logger
	startTime time.Time
}

// NewBlockchain создает новый экземпляр блокчейна-заглушки
func NewBlockchain(config *config.Config, logger *logrus.Logger) (*Blockchain, error) {
	bc := &Blockchain{
		config:    config,
		logger:    logger,
		startTime: time.Now().UTC(),
	}
	
	logger.Info("Created blockchain stub implementation")
	
	return bc, nil
}

// Start запускает блокчейн (заглушка)
func (bc *Blockchain) Start() {
	bc.logger.Info("Starting blockchain stub implementation")
}

// GetStats возвращает статистику блокчейна (заглушка)
func (bc *Blockchain) GetStats() models.BlockchainStats {
	return models.BlockchainStats{
		Height:            0,
		LastBlockTime:     bc.startTime,
		TotalTransactions: 0,
		PendingTxs:        0,
	}
}

// HandleBlockchainMessage обрабатывает сообщение (заглушка)
func (bc *Blockchain) HandleBlockchainMessage(message *models.GossipMessage) error {
	// Просто логируем полученное сообщение без фактической обработки
	messageJson, _ := json.MarshalIndent(message, "", "  ")
	bc.logger.Infof("Received message:\nType: %s\nMessageID: %s\nOrigin: %s\nTimestamp: %s\nPayload: %s", 
		message.MessageType, 
		message.MessageID, 
		message.OriginID, 
		message.Timestamp.Format(time.RFC3339),
		string(messageJson))
	
	return nil
}