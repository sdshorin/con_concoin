package hooks

import (
	"concoin/conrun/pkg/models"

	"github.com/sirupsen/logrus"
)

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

// HookManager управляет всеми хуками
type HookManager struct {
	hooks   []Hook
	logger  *logrus.Logger
	rootDir string
}

// NewHookManager создает новый менеджер хуков
func NewHookManager(rootDir string, logger *logrus.Logger) *HookManager {
	return &HookManager{
		hooks:   make([]Hook, 0),
		logger:  logger,
		rootDir: rootDir,
	}
}

// AddHook добавляет новый хук
func (hm *HookManager) AddHook(hook Hook) {
	hm.hooks = append(hm.hooks, hook)
}

// ValidateMessage проверяет валидность сообщения через все подходящие хуки
func (hm *HookManager) ValidateMessage(message *models.GossipMessage, msgType MessageType) bool {
	// Проверяем, есть ли хотя бы один хук, который должен обрабатывать это сообщение
	hasHandler := false
	for _, hook := range hm.hooks {
		if hook.ShouldHandle(message.MessageType) {
			hasHandler = true
			break
		}
	}

	if !hasHandler {
		hm.logger.Infof("HookManager: ValidateMessage: no handler for message: %s", message.MessageID)
		return false
	}

	// Проверяем валидность через все подходящие хуки
	for _, hook := range hm.hooks {
		if hook.ShouldHandle(message.MessageType) {
			if hook.Validate(message, msgType) {
				return true // Достаточно одного валидного хука
			}
		}
	}

	return false
}

// ProcessMessage обрабатывает сообщение через все подходящие хуки
func (hm *HookManager) ProcessMessage(message *models.GossipMessage, msgType MessageType) bool {
	// Проверяем, есть ли хотя бы один хук, который должен обрабатывать это сообщение
	hasHandler := false
	for _, hook := range hm.hooks {
		if hook.ShouldHandle(message.MessageType) {
			hasHandler = true
			break
		}
	}

	if !hasHandler {
		hm.logger.Infof("HookManager: ProcessMessage: no handler for message: %s", message.MessageID)
		return false
	}

	// Проверяем валидность через все подходящие хуки
	isValid := false
	for _, hook := range hm.hooks {
		if hook.ShouldHandle(message.MessageType) {
			if hook.Validate(message, msgType) {
				isValid = true
				// Обрабатываем сообщение
				if err := hook.Handle(message, msgType); err != nil {
					hm.logger.Errorf("Failed to handle message with hook: %v", err)
				}
			}
		}
	}

	return isValid
}
