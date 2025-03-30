package hooks

import (
	"concoin/conrun/pkg/models"

	"github.com/sirupsen/logrus"
)

// DebugHook представляет собой хук для отладки
type DebugHook struct {
	logger *logrus.Logger
}

// NewDebugHook создает новый дебажный хук
func NewDebugHook(logger *logrus.Logger) *DebugHook {
	return &DebugHook{
		logger: logger,
	}
}

// ShouldHandle проверяет, должен ли хук обрабатывать сообщение
func (h *DebugHook) ShouldHandle(messageType string) bool {
	return messageType == "user_message"
}

// Validate проверяет валидность сообщения
func (h *DebugHook) Validate(message *models.GossipMessage, msgType MessageType) bool {
	// Для дебажного хука все сообщения валидны
	return true
}

// Handle обрабатывает сообщение
func (h *DebugHook) Handle(message *models.GossipMessage, msgType MessageType) error {
	h.logger.Infof("Debug hook received message: Type=%s, ID=%s, Origin=%s, Payload=%v",
		message.MessageType,
		message.MessageID,
		message.OriginID,
		message.Payload)
	return nil
}
