package gossip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/hooks"
	"concoin/conrun/pkg/models"
	"concoin/conrun/pkg/storage"

	"github.com/sirupsen/logrus"
)

// GossipProtocol представляет собой реализацию Gossip протокола
type GossipProtocol struct {
	config         *config.Config
	messageHistory map[string]time.Time // Хеш-таблица для хранения истории сообщений
	historyMutex   sync.RWMutex
	peerList       []models.Peer
	peerMutex      sync.RWMutex
	hookManager    *hooks.HookManager
	logger         *logrus.Logger
	storage        *storage.Storage
}

// NewGossipProtocol создает новый экземпляр Gossip протокола
func NewGossipProtocol(config *config.Config, logger *logrus.Logger, storage *storage.Storage, hookManager *hooks.HookManager) *GossipProtocol {
	return &GossipProtocol{
		config:         config,
		messageHistory: make(map[string]time.Time),
		logger:         logger,
		storage:        storage,
		hookManager:    hookManager,
	}
}

// UpdatePeers обновляет список пиров
func (g *GossipProtocol) UpdatePeers(peers []models.Peer) {
	g.peerMutex.Lock()
	defer g.peerMutex.Unlock()
	g.peerList = peers
}

// Start запускает протокол
func (g *GossipProtocol) Start() {
	g.logger.Info("Starting Gossip protocol")

	// Периодически отправляем сообщения
	go func() {
		ticker := time.NewTicker(g.config.GossipConfig.SyncInterval)
		defer ticker.Stop()

		for range ticker.C {
			// Очищаем историю сообщений от старых записей
			g.cleanMessageHistory()
		}
	}()
}

// BroadcastMessage отправляет сообщение всем пирам
func (g *GossipProtocol) BroadcastMessage(messageType string, payload interface{}) error {
	// Генерируем случайный ID сообщения
	messageID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Создаем сообщение Gossip
	message := &models.GossipMessage{
		MessageID:   messageID,
		OriginID:    g.config.NodeID,
		Timestamp:   time.Now().UTC(),
		TTL:         g.config.GossipConfig.MessageTTL,
		MessageType: messageType,
		Payload:     payload,
	}

	// Добавляем сообщение в историю
	g.addToMessageHistory(message.MessageID)

	// Отправляем сообщение случайным пирам
	return g.spreadMessage(message)
}

// HandleMessage обрабатывает входящее сообщение
func (g *GossipProtocol) HandleMessage(message *models.GossipMessage) error {
	// Проверяем TTL
	if message.TTL <= 0 {
		g.logger.Debugf("Message TTL expired: %s", message.MessageID)
		return nil
	}

	// Проверяем время сообщения
	if time.Since(message.Timestamp) > g.config.GossipConfig.MessageMaxAge {
		g.logger.Debugf("Ignoring old message: %s", message.MessageID)
		return nil
	}

	// Проверяем, не обрабатывали ли мы уже это сообщение
	if g.isMessageProcessed(message.MessageID) {
		g.logger.Debugf("Message already processed: %s", message.MessageID)
		return nil
	}

	// Если сообщения нет в истории, проверяем его в storage
	if !g.isMessageProcessed(message.MessageID) {
		// Проверяем, не истек ли срок действия сообщения
		isExpired, err := g.storage.IsMessageExpired(message.MessageID, g.config.GossipConfig.MessageMaxAge)
		if err != nil {
			g.logger.Warnf("Failed to check message expiration: %v", err)
			return err
		}
		if isExpired {
			g.logger.Debugf("Message expired in storage: %s", message.MessageID)
			return nil
		}
	}

	// Проверяем валидность сообщения через хуки
	if !g.hookManager.ValidateMessage(message, hooks.MessageTypePull) {
		g.logger.Warnf("Message validation failed: %s", message.MessageID)
		return fmt.Errorf("message validation failed: %s", message.MessageID)
	}

	// Добавляем сообщение в историю
	g.addToMessageHistory(message.MessageID)

	// Уменьшаем TTL и передаем сообщение дальше
	message.TTL--
	if err := g.spreadMessage(message); err != nil {
		g.logger.Warnf("Failed to spread message: %v", err)
	}

	// Обрабатываем сообщение через хуки
	g.hookManager.ProcessMessage(message, hooks.MessageTypePull)

	return nil
}

// spreadMessage отправляет сообщение случайным пирам
func (g *GossipProtocol) spreadMessage(message *models.GossipMessage) error {
	g.peerMutex.RLock()
	defer g.peerMutex.RUnlock()

	if len(g.peerList) == 0 {
		g.logger.Debug("No peers to spread message to")
		return nil
	}

	// Выбираем случайные пиры
	selectedPeers := g.selectRandomPeers(g.config.GossipConfig.BranchingFactor)

	// Отправляем сообщение выбранным пирам
	var wg sync.WaitGroup
	for _, peer := range selectedPeers {
		wg.Add(1)
		go func(p models.Peer) {
			defer wg.Done()
			if err := g.sendMessageToPeer(message, p); err != nil {
				g.logger.Warnf("Failed to send message to peer %s: %v", p.NodeID, err)
			}
		}(peer)
	}

	wg.Wait()
	return nil
}

// selectRandomPeers выбирает случайные пиры
func (g *GossipProtocol) selectRandomPeers(count int) []models.Peer {
	if len(g.peerList) <= count {
		return g.peerList
	}

	// Создаем случайную выборку
	indices := rand.Perm(len(g.peerList))[:count]
	selectedPeers := make([]models.Peer, count)

	for i, idx := range indices {
		selectedPeers[i] = g.peerList[idx]
	}

	return selectedPeers
}

// sendMessageToPeer отправляет сообщение конкретному пиру
func (g *GossipProtocol) sendMessageToPeer(message *models.GossipMessage, peer models.Peer) error {
	// Сериализуем сообщение
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Отправляем HTTP запрос
	url := fmt.Sprintf("http://%s/gossip", peer.Address)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received bad status code: %d", resp.StatusCode)
	}

	return nil
}

// addToMessageHistory добавляет сообщение в историю
func (g *GossipProtocol) addToMessageHistory(messageID string) {
	g.historyMutex.Lock()
	defer g.historyMutex.Unlock()

	// Проверяем размер истории
	if len(g.messageHistory) >= g.config.GossipConfig.HistorySize {
		// Удаляем самое старое сообщение
		var oldestID string
		var oldestTime time.Time

		for id, t := range g.messageHistory {
			if oldestID == "" || t.Before(oldestTime) {
				oldestID = id
				oldestTime = t
			}
		}

		delete(g.messageHistory, oldestID)
	}

	// Добавляем новое сообщение
	g.messageHistory[messageID] = time.Now()
}

// isMessageProcessed проверяет, обрабатывали ли мы уже это сообщение
func (g *GossipProtocol) isMessageProcessed(messageID string) bool {
	g.historyMutex.RLock()
	defer g.historyMutex.RUnlock()

	_, exists := g.messageHistory[messageID]
	if exists {
		return true
	}

	return g.storage.HasMessage(messageID)
}

// cleanMessageHistory удаляет старые сообщения из истории
func (g *GossipProtocol) cleanMessageHistory() {
	g.historyMutex.Lock()
	defer g.historyMutex.Unlock()

	now := time.Now()
	for id, t := range g.messageHistory {
		if now.Sub(t) > g.config.GossipConfig.MessageMaxAge {
			delete(g.messageHistory, id)
		}
	}
}
