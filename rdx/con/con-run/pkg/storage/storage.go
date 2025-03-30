package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"concoin/conrun/pkg/models"
)

// Storage обеспечивает хранение данных узла
type Storage struct {
	dataDir    string
	peersMutex sync.RWMutex
}

// NewStorage создает новый экземпляр хранилища
func NewStorage(dataDir string) *Storage {
	// Создаем директорию для сообщений
	messagesDir := filepath.Join(dataDir, "messages")
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		// В MVP просто игнорируем ошибку
	}

	return &Storage{
		dataDir: dataDir,
	}
}

// SavePeer сохраняет информацию о пире
func (s *Storage) SavePeer(peer *models.Peer) error {
	s.peersMutex.Lock()
	defer s.peersMutex.Unlock()

	peerData, err := json.MarshalIndent(peer, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal peer: %w", err)
	}

	peerPath := filepath.Join(s.dataDir, "peers", fmt.Sprintf("%s.json", peer.NodeID))
	if err := os.WriteFile(peerPath, peerData, 0644); err != nil {
		return fmt.Errorf("failed to save peer to file: %w", err)
	}

	return nil
}

// GetPeers получает всех известных пиров
func (s *Storage) GetPeers() ([]*models.Peer, error) {
	s.peersMutex.RLock()
	defer s.peersMutex.RUnlock()

	peersDir := filepath.Join(s.dataDir, "peers")
	entries, err := os.ReadDir(peersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Peer{}, nil
		}
		return nil, fmt.Errorf("failed to read peers directory: %w", err)
	}

	var peers []*models.Peer
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		peerPath := filepath.Join(peersDir, entry.Name())
		peerData, err := os.ReadFile(peerPath)
		if err != nil {
			continue
		}

		var peer models.Peer
		if err := json.Unmarshal(peerData, &peer); err != nil {
			continue
		}

		peers = append(peers, &peer)
	}

	return peers, nil
}

// SaveMessage сохраняет сообщение в хранилище
func (s *Storage) SaveMessage(message *models.GossipMessage) error {
	messageData, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	messagesDir := filepath.Join(s.dataDir, "messages")
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create messages directory: %w", err)
	}

	messagePath := filepath.Join(messagesDir, fmt.Sprintf("%s.json", message.MessageID))
	if err := os.WriteFile(messagePath, messageData, 0644); err != nil {
		return fmt.Errorf("failed to save message to file: %w", err)
	}

	return nil
}

// GetMessage получает сообщение по ID
func (s *Storage) GetMessage(messageID string) (*models.GossipMessage, error) {
	messagePath := filepath.Join(s.dataDir, "messages", fmt.Sprintf("%s.json", messageID))
	messageData, err := os.ReadFile(messagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read message file: %w", err)
	}

	var message models.GossipMessage
	if err := json.Unmarshal(messageData, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// GetMessageList получает список всех сообщений
func (s *Storage) GetMessageList() ([]string, error) {
	messagesDir := filepath.Join(s.dataDir, "messages")
	entries, err := os.ReadDir(messagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read messages directory: %w", err)
	}

	var messageIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Убираем расширение .json
		messageID := strings.TrimSuffix(entry.Name(), ".json")
		messageIDs = append(messageIDs, messageID)
	}

	return messageIDs, nil
}

// HasMessage проверяет наличие сообщения
func (s *Storage) HasMessage(messageID string) bool {
	messagePath := filepath.Join(s.dataDir, "messages", fmt.Sprintf("%s.json", messageID))
	_, err := os.Stat(messagePath)
	return err == nil
}

// IsMessageExpired проверяет, истек ли срок действия сообщения
func (s *Storage) IsMessageExpired(messageID string, maxAge time.Duration) (bool, error) {
	message, err := s.GetMessage(messageID)
	if err != nil {
		return false, err
	}

	return time.Since(message.Timestamp) > maxAge, nil
}
