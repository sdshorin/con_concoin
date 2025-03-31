package hooks

import (
	"concoin/conrun/pkg/interfaces"

	"concoin/conrun/pkg/models"

	"github.com/sirupsen/logrus"
)

// HookManager управляет всеми хуками
type HookManager struct {
	hooks   []interfaces.Hook
	logger  *logrus.Logger
	rootDir string
}

// NewHookManager создает новый менеджер хуков
func NewHookManager(rootDir string, logger *logrus.Logger) *HookManager {
	return &HookManager{
		hooks:   make([]interfaces.Hook, 0),
		logger:  logger,
		rootDir: rootDir,
	}
}

// AddHook добавляет новый хук
func (hm *HookManager) AddHook(hook interfaces.Hook) {
	hm.hooks = append(hm.hooks, hook)
}

// ValidateMessage проверяет валидность сообщения через все подходящие хуки
func (hm *HookManager) ValidateMessage(message *models.GossipMessage, msgType interfaces.MessageType) bool {
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
func (hm *HookManager) ProcessMessage(message *models.GossipMessage, msgType interfaces.MessageType) bool {
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
