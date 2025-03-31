package interfaces

import (
	"concoin/conrun/pkg/models"
)

// GossipProtocolInterface определяет интерфейс для Gossip протокола
type GossipProtocolInterface interface {
	Start()
	UpdatePeers(peers []models.Peer)
	HandleMessage(message *models.GossipMessage) error
}

// PexProtocolInterface определяет интерфейс для PEX протокола
type PexProtocolInterface interface {
	Start()
	SetOnPeersListHandler(handler func(peers []models.Peer))
	AddPeer(peer models.Peer) bool
	GetPeers() []models.Peer
	HandlePexRequest(request models.PexMessage) models.PexMessage
}

// StorageInterface определяет интерфейс для хранилища
type StorageInterface interface {
	SavePeer(peer *models.Peer) error
	GetPeers() ([]*models.Peer, error)
	SaveMessage(message *models.GossipMessage) error
	GetMessage(messageID string) (*models.GossipMessage, error)
	GetMessageList() ([]string, error)
	HasMessage(messageID string) bool
}

// MessageType представляет тип сообщения
type MessageType string

const (
	MessageTypeLoaded MessageType = "loaded" // Сообщение загружено с диска
	MessageTypePull   MessageType = "pull"   // Сообщение получено от другой ноды
	MessageTypePush   MessageType = "push"   // Сообщение отправлено другой ноде
)

// Hook представляет собой интерфейс для обработчиков сообщений
type Hook interface {
	// ShouldHandle проверяет, должен ли хук обрабатывать сообщение данного типа
	ShouldHandle(messageType string) bool
	// Validate проверяет валидность сообщения
	Validate(message *models.GossipMessage, msgType MessageType) bool
	// Handle обрабатывает валидное сообщение
	Handle(message *models.GossipMessage, msgType MessageType) error
}

// HookManagerInterface определяет интерфейс для менеджера хуков
type HookManagerInterface interface {
	ValidateMessage(message *models.GossipMessage, msgType MessageType) bool
	ProcessMessage(message *models.GossipMessage, msgType MessageType) bool
	AddHook(hook Hook)
}
