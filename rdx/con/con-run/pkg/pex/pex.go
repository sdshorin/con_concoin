package pex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/interfaces"
	"concoin/conrun/pkg/models"

	"github.com/sirupsen/logrus"
)

// PexProtocol представляет собой реализацию PEX протокола
type PexProtocol struct {
	config      *config.Config
	peerTable   map[string]models.Peer // nodeId -> peer
	storage     interfaces.StorageInterface
	mutex       sync.RWMutex
	logger      *logrus.Logger
	hookManager interfaces.HookManagerInterface
	onPeersList func(peers []models.Peer)
}

// NewPexProtocol создает новый экземпляр PEX протокола
func NewPexProtocol(config *config.Config, storage interfaces.StorageInterface, logger *logrus.Logger, hookManager interfaces.HookManagerInterface) *PexProtocol {
	return &PexProtocol{
		config:      config,
		peerTable:   make(map[string]models.Peer),
		storage:     storage,
		logger:      logger,
		hookManager: hookManager,
	}
}

// SetOnPeersListHandler устанавливает обработчик для обновления списка пиров
func (p *PexProtocol) SetOnPeersListHandler(handler func(peers []models.Peer)) {
	p.onPeersList = handler
}

// Start запускает протокол
func (p *PexProtocol) Start() {
	p.logger.Info("Starting PEX protocol")

	// Загружаем известных пиров из хранилища
	p.loadPeersFromStorage()

	// Добавляем seed-узлы, если список пиров пуст
	if len(p.peerTable) == 0 {
		p.logger.Info("No peers found, adding seed nodes")
		p.addSeedNodes()
	}

	p.logger.Infof("Initial peer table size: %d", len(p.peerTable))

	// Периодически обмениваемся пирами
	go p.exchangeLoop()

	// Периодически очищаем неактивных пиров
	go p.cleanupLoop()
}

// loadPeersFromStorage загружает пиров из хранилища
func (p *PexProtocol) loadPeersFromStorage() {
	peers, err := p.storage.GetPeers()
	if err != nil {
		p.logger.Warnf("Failed to load peers from storage: %v", err)
		return
	}

	p.logger.Infof("Loaded %d peers from storage", len(peers))

	p.mutex.Lock()
	defer p.mutex.Unlock()

	activePeers := 0
	for _, peer := range peers {
		// Пропускаем пиров с истекшим временем жизни
		if time.Since(peer.LastSeen) > p.config.PexConfig.PeerTTL {
			p.logger.Debugf("Skipping expired peer: %s (last seen: %v)", peer.NodeID, peer.LastSeen)
			continue
		}

		p.peerTable[peer.NodeID] = *peer
		activePeers++
	}

	p.logger.Infof("Added %d active peers to peer table", activePeers)
	p.notifyPeersUpdated()
}

// addSeedNodes добавляет seed-узлы в таблицу пиров
func (p *PexProtocol) addSeedNodes() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.logger.Infof("Adding seed nodes: %v", p.config.SeedNodes)

	addedSeeds := 0
	for _, seedNode := range p.config.SeedNodes {
		// Пропускаем свой собственный адрес
		if seedNode == fmt.Sprintf("127.0.0.1:%d", p.config.Port) {
			p.logger.Debugf("Skipping self address: %s", seedNode)
			continue
		}

		nodeID := fmt.Sprintf("seed-%s", seedNode)
		peer := models.Peer{
			NodeID:   nodeID,
			Address:  seedNode,
			LastSeen: time.Now(),
		}

		p.peerTable[nodeID] = peer
		addedSeeds++
		p.logger.Infof("Added seed node: %s", seedNode)
	}

	p.logger.Infof("Added %d seed nodes to peer table", addedSeeds)
	p.notifyPeersUpdated()
}

// AddPeer добавляет новый пир в таблицу
func (p *PexProtocol) AddPeer(peer models.Peer) bool {
	p.logger.Debugf("Attempting to add peer: %s (%s)", peer.NodeID, peer.Address)

	// Пропускаем свой собственный адрес
	if peer.Address == fmt.Sprintf("127.0.0.1:%d", p.config.Port) {
		p.logger.Debugf("Skipping self address: %s", peer.Address)
		return false
	}

	// Проверяем формат адреса
	if !p.isValidAddress(peer.Address) {
		p.logger.Warnf("Invalid peer address format: %s", peer.Address)
		return false
	}

	// Проверяем доступность пира
	if !p.testConnection(peer.Address) {
		p.logger.Debugf("Peer is not reachable: %s", peer.Address)
		return false
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Пропускаем, если уже добавлен и недавно обновлен
	existingPeer, exists := p.peerTable[peer.NodeID]
	if exists {
		if existingPeer.LastSeen.After(peer.LastSeen) {
			p.logger.Debugf("Skipping peer %s: existing peer is newer", peer.NodeID)
			return false
		}
		p.logger.Infof("Updating existing peer: %s", peer.NodeID)
	}

	// Ограничиваем размер таблицы пиров
	if len(p.peerTable) >= p.config.PexConfig.MaxPeers && !exists {
		// Удаляем самого старого пира
		var oldestID string
		var oldestTime time.Time

		for id, p := range p.peerTable {
			if oldestID == "" || p.LastSeen.Before(oldestTime) {
				oldestID = id
				oldestTime = p.LastSeen
			}
		}

		p.logger.Infof("Removing oldest peer: %s (last seen: %v)", oldestID, oldestTime)
		delete(p.peerTable, oldestID)
	}

	// Добавляем нового пира
	p.peerTable[peer.NodeID] = peer
	p.logger.Infof("Added new peer: %s (%s)", peer.NodeID, peer.Address)

	// Сохраняем пира в хранилище
	go p.storage.SavePeer(&peer)

	// Запускаем синхронизацию с новым пиром
	go func() {
		if err := p.syncMessagesWithPeer(peer); err != nil {
			p.logger.Warnf("Failed to sync messages with peer %s: %v", peer.NodeID, err)
		}
	}()

	p.notifyPeersUpdated()
	return true
}

// exchangeLoop периодически обменивается пирами
func (p *PexProtocol) exchangeLoop() {
	// Определяем интервал обмена
	interval := p.config.PexConfig.ExchangeInterval

	p.logger.Info("Starting peer exchange loop")

	ticker := time.NewTicker(1)
	defer ticker.Stop()

	for range ticker.C {
		// Проверяем количество известных пиров
		p.mutex.RLock()
		peerCount := len(p.peerTable)
		p.mutex.RUnlock()

		p.logger.Debugf("Current peer count: %d", peerCount)

		// Если пиров мало, увеличиваем частоту обмена
		if peerCount < p.config.PexConfig.LowConnectivityThreshold {
			p.logger.Info("Low peer count, increasing exchange frequency")
			interval = p.config.PexConfig.ExchangeInterval / 5
		} else {
			interval = p.config.PexConfig.ExchangeInterval
		}

		// Отправляем запросы на обмен пирами
		p.exchangePeers()

		// Обновляем ticker, если интервал изменился
		ticker.Reset(interval)
	}
}

// exchangePeers отправляет PEX запросы случайным пирам
func (p *PexProtocol) exchangePeers() {
	p.mutex.RLock()
	if len(p.peerTable) == 0 {
		p.mutex.RUnlock()
		p.logger.Warn("No peers available for exchange")
		return
	}

	// Выбираем случайного пира
	peers := make([]models.Peer, 0, len(p.peerTable))
	for _, peer := range p.peerTable {
		peers = append(peers, peer)
	}
	p.mutex.RUnlock()

	// Выбираем случайного пира
	if len(peers) == 0 {
		p.logger.Warn("No peers available for exchange")
		return
	}

	targetPeer := peers[rand.Intn(len(peers))]
	p.logger.Infof("Selected peer for exchange: %s (%s)", targetPeer.NodeID, targetPeer.Address)

	// Создаем PEX запрос
	request := models.PexMessage{
		MessageID: fmt.Sprintf("pex-req-%d", time.Now().UnixNano()),
		Type:      models.PexRequest,
		Timestamp: time.Now().UTC(),
		Peers:     []models.Peer{},
	}

	// Отправляем запрос
	p.sendPexRequest(targetPeer, request)
}

// sendPexRequest отправляет PEX запрос пиру
func (p *PexProtocol) sendPexRequest(peer models.Peer, request models.PexMessage) {
	p.logger.Debugf("Sending PEX request to %s (%s)", peer.NodeID, peer.Address)

	// Добавляем информацию о себе в запрос
	selfPeer := models.Peer{
		NodeID:   p.config.NodeID,
		Address:  fmt.Sprintf("127.0.0.1:%d", p.config.Port),
		LastSeen: time.Now(),
	}
	request.Peers = append(request.Peers, selfPeer)

	// Сериализуем запрос
	data, err := json.Marshal(request)
	if err != nil {
		p.logger.Warnf("Failed to marshal PEX request: %v", err)
		return
	}

	// Отправляем HTTP запрос
	url := fmt.Sprintf("http://%s/pex", peer.Address)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		p.logger.Warnf("Failed to send PEX request to %s: %v", peer.Address, err)
		return
	}
	defer resp.Body.Close()

	// Обновляем время последнего обращения к пиру
	p.updatePeerLastSeen(peer.NodeID)

	// Обрабатываем ответ
	if resp.StatusCode != http.StatusOK {
		p.logger.Warnf("Received bad status code from PEX request: %d", resp.StatusCode)
		return
	}

	// Декодируем ответ
	var response models.PexMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		p.logger.Warnf("Failed to decode PEX response: %v", err)
		return
	}

	p.logger.Infof("Received PEX response from %s with %d peers", peer.NodeID, len(response.Peers))

	// Обрабатываем полученных пиров
	addedPeers := 0
	for _, receivedPeer := range response.Peers {
		if p.AddPeer(receivedPeer) {
			addedPeers++
		}
	}

	p.logger.Infof("Added %d new peers from PEX response", addedPeers)
}

// HandlePexRequest обрабатывает входящий PEX запрос
func (p *PexProtocol) HandlePexRequest(request models.PexMessage) models.PexMessage {
	p.logger.Infof("Received PEX request with %d peers", len(request.Peers))

	// Обновляем информацию об отправителе, если она есть
	if len(request.Peers) > 0 {
		addedPeers := 0
		for _, peer := range request.Peers {
			if p.AddPeer(peer) {
				addedPeers++
			}
		}
		p.logger.Infof("Added %d new peers from PEX request", addedPeers)
	}

	// Создаем ответ
	response := models.PexMessage{
		MessageID: fmt.Sprintf("pex-res-%d", time.Now().UnixNano()),
		Type:      models.PexResponse,
		Timestamp: time.Now().UTC(),
		Peers:     p.getRandomPeers(p.config.PexConfig.MaxPeersPerExchange),
	}

	// Добавляем информацию о себе в ответ
	selfPeer := models.Peer{
		NodeID:   p.config.NodeID,
		Address:  fmt.Sprintf("127.0.0.1:%d", p.config.Port),
		LastSeen: time.Now(),
	}
	response.Peers = append(response.Peers, selfPeer)

	p.logger.Infof("Sending PEX response with %d peers", len(response.Peers))
	return response
}

// getRandomPeers возвращает случайных пиров из таблицы
func (p *PexProtocol) getRandomPeers(count int) []models.Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Собираем всех активных пиров
	activePeers := make([]models.Peer, 0, len(p.peerTable))
	for _, peer := range p.peerTable {
		// Отбираем только недавно активных пиров
		if time.Since(peer.LastSeen) < 30*time.Minute {
			activePeers = append(activePeers, peer)
		}
	}

	// Если пиров меньше, чем запрошено
	if len(activePeers) <= count {
		return activePeers
	}

	// Выбираем случайных пиров
	result := make([]models.Peer, 0, count)
	indices := rand.Perm(len(activePeers))

	for i := 0; i < count; i++ {
		result = append(result, activePeers[indices[i]])
	}

	return result
}

// cleanupLoop периодически удаляет неактивных пиров
func (p *PexProtocol) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanupInactivePeers()
	}
}

// cleanupInactivePeers удаляет неактивных пиров
func (p *PexProtocol) cleanupInactivePeers() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	for id, peer := range p.peerTable {
		if now.Sub(peer.LastSeen) > p.config.PexConfig.PeerTTL {
			delete(p.peerTable, id)
		}
	}

	p.notifyPeersUpdated()
}

// updatePeerLastSeen обновляет время последнего обращения к пиру
func (p *PexProtocol) updatePeerLastSeen(nodeID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if peer, exists := p.peerTable[nodeID]; exists {
		peer.LastSeen = time.Now()
		p.peerTable[nodeID] = peer

		// Асинхронно сохраняем в хранилище
		peerCopy := peer
		go p.storage.SavePeer(&peerCopy)
	}
}

// GetPeers возвращает список всех известных пиров
func (p *PexProtocol) GetPeers() []models.Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	peers := make([]models.Peer, 0, len(p.peerTable))
	for _, peer := range p.peerTable {
		peers = append(peers, peer)
	}

	return peers
}

// isValidAddress проверяет формат адреса
func (p *PexProtocol) isValidAddress(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}

	// Проверяем хост
	if net.ParseIP(host) == nil && host != "localhost" {
		return false
	}

	// Проверяем порт
	if _, err := net.LookupPort("tcp", port); err != nil {
		return false
	}

	return true
}

// testConnection проверяет доступность пира
func (p *PexProtocol) testConnection(address string) bool {
	// Простая проверка доступности
	url := fmt.Sprintf("http://%s/ping", address)
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// notifyPeersUpdated уведомляет об обновлении списка пиров
func (p *PexProtocol) notifyPeersUpdated() {
	if p.onPeersList != nil {
		peers := make([]models.Peer, 0, len(p.peerTable))
		for _, peer := range p.peerTable {
			peers = append(peers, peer)
		}

		p.onPeersList(peers)
	}
}

// syncMessagesWithPeer синхронизирует сообщения с новым пиром
func (p *PexProtocol) syncMessagesWithPeer(peer models.Peer) error {
	p.logger.Infof("Starting message sync with peer %s", peer.NodeID)

	// Получаем список локальных сообщений
	localMessageIDs, err := p.storage.GetMessageList()
	if err != nil {
		return fmt.Errorf("failed to get local message list: %w", err)
	}

	// Создаем карту локальных сообщений для быстрого поиска
	localMessages := make(map[string]bool)
	for _, msgID := range localMessageIDs {
		localMessages[msgID] = true
	}

	// Получаем список сообщений пира
	url := fmt.Sprintf("http://%s/messages", peer.Address)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get message list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var messageIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&messageIDs); err != nil {
		return fmt.Errorf("failed to decode message list: %w", err)
	}

	// Загружаем отсутствующие сообщения
	var syncMessages int
	for _, msgID := range messageIDs {
		if !localMessages[msgID] {
			if err := p.downloadMessage(peer, msgID); err != nil {
				p.logger.Warnf("Failed to download message %s: %v", msgID, err)
				continue
			}
			syncMessages++
		}
	}
	p.logger.Infof("Synced %d messages with peer %s", syncMessages, peer.NodeID)

	return nil
}

// downloadMessage загружает конкретное сообщение от пира
func (p *PexProtocol) downloadMessage(peer models.Peer, messageID string) error {
	url := fmt.Sprintf("http://%s/messages/%s", peer.Address, messageID)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var message models.GossipMessage
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	// Проверяем валидность сообщения через хуки
	if !p.hookManager.ValidateMessage(&message, interfaces.MessageTypeLoaded) {
		p.logger.Warnf("Message validation failed during download: %s", messageID)
		return fmt.Errorf("message validation failed: %s", messageID)
	}

	// Сохраняем сообщение
	if err := p.storage.SaveMessage(&message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Обрабатываем сообщение через хуки
	p.hookManager.ProcessMessage(&message, interfaces.MessageTypeLoaded)

	return nil
}
