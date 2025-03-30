package hooks

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	"concoin/conrun/pkg/models"

	"github.com/sirupsen/logrus"
)

// BlockchainHook представляет собой хук для обработки сообщений блокчейна
type BlockchainHook struct {
	logger     *logrus.Logger
	rootDir    string
	scriptsDir string
}

// NewBlockchainHook создает новый хук для блокчейна
func NewBlockchainHook(rootDir string, logger *logrus.Logger) *BlockchainHook {
	// Получаем абсолютный путь к директории скриптов
	scriptsDir := filepath.Join("pkg", "hooks", "blockchain_tools")
	absScriptsDir, err := filepath.Abs(scriptsDir)
	if err != nil {
		logger.Errorf("Failed to get absolute path to scripts: %v", err)
	}

	logger.Infof("BlockchainHook: Root directory: %s", rootDir)
	logger.Infof("BlockchainHook: Scripts directory: %s", absScriptsDir)

	return &BlockchainHook{
		logger:     logger,
		rootDir:    rootDir,
		scriptsDir: absScriptsDir,
	}
}

// ShouldHandle проверяет, должен ли хук обрабатывать сообщение
func (h *BlockchainHook) ShouldHandle(messageType string) bool {
	h.logger.Infof("BlockchainHook: ShouldHandle: %s", messageType)
	return messageType == "blockchain_concoin"
}

func (h *BlockchainHook) save_temp_message(message *models.GossipMessage, messagesDir string) (string, error) {
	// Сохраняем сообщение во временный файл
	messageFile := filepath.Join(messagesDir, fmt.Sprintf("%s_%d.json", message.MessageID, rand.Intn(1000000)))
	messageData, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		h.logger.Errorf("Failed to marshal message: %v", err)
		return "", err
	}

	if err := os.WriteFile(messageFile, messageData, 0644); err != nil {
		h.logger.Errorf("Failed to write message file: %v", err)
		return "", err
	}

	return messageFile, nil
}

// Validate проверяет валидность сообщения
func (h *BlockchainHook) Validate(message *models.GossipMessage, msgType MessageType) bool {
	h.logger.Infof("BlockchainHook: Validate start: %s", message.MessageID)
	// Создаем директорию для сообщений, если её нет
	messagesDir := filepath.Join(h.rootDir, "messages_to_check")
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		h.logger.Errorf("Failed to create messages directory: %v", err)
		return false
	}

	messageFile, err := h.save_temp_message(message, messagesDir)
	if err != nil {
		h.logger.Errorf("Failed to save message: %v", err)
		return false
	} else {
		h.logger.Infof("Message saved to: %s", messageFile)
	}

	// Запускаем coin_valid
	cmd := exec.Command(filepath.Join(h.scriptsDir, "coin_valid"), messageFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.logger.Errorf("coin_valid failed: %v\nOutput: %s", err, string(output))
		os.Remove(messageFile)
		return false
	}
	h.logger.Infof("coin_valid output: %s", string(output))
	h.logger.Infof("BlockchainHook: Validate end - valid: %s", message.MessageID)
	os.Remove(messageFile)
	return true
}

// Handle обрабатывает сообщение
func (h *BlockchainHook) Handle(message *models.GossipMessage, msgType MessageType) error {
	h.logger.Infof("BlockchainHook: Handle start: %s", message.MessageID)

	messagesDir := filepath.Join(h.rootDir, "messages_to_check")
	messageFile, err := h.save_temp_message(message, messagesDir)
	if err != nil {
		h.logger.Errorf("Failed to save message: %v", err)
		return err
	}

	// Запускаем process_message
	cmd := exec.Command(filepath.Join(h.scriptsDir, "process_message"), messageFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.logger.Errorf("process_message failed: %v\nOutput: %s", err, string(output))
		return err
	}
	h.logger.Infof("process_message output: %s", string(output))
	h.logger.Infof("BlockchainHook: Handle end: %s", message.MessageID)
	// Удаляем временный файл
	os.Remove(messageFile)
	return nil
}
