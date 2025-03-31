package tests

import (
	"os"
	"testing"
	"time"

	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/gossip"
	"concoin/conrun/pkg/interfaces"
	"concoin/conrun/pkg/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// MockHookManager mocks the hook manager for testing
type MockHookManager struct {
	mock.Mock
}

func (m *MockHookManager) ValidateMessage(message *models.GossipMessage, msgType interfaces.MessageType) bool {
	args := m.Called(message, msgType)
	return args.Bool(0)
}

func (m *MockHookManager) ProcessMessage(message *models.GossipMessage, msgType interfaces.MessageType) bool {
	args := m.Called(message, msgType)
	return args.Bool(0)
}

func (m *MockHookManager) AddHook(hook interfaces.Hook) {
	m.Called(hook)
}

// MockStorage mocks the storage for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SavePeer(peer *models.Peer) error {
	args := m.Called(peer)
	return args.Error(0)
}

func (m *MockStorage) GetPeers() ([]*models.Peer, error) {
	args := m.Called()
	return args.Get(0).([]*models.Peer), args.Error(1)
}

func (m *MockStorage) SaveMessage(message *models.GossipMessage) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockStorage) GetMessage(messageID string) (*models.GossipMessage, error) {
	args := m.Called(messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GossipMessage), args.Error(1)
}

func (m *MockStorage) GetMessageList() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) HasMessage(messageID string) bool {
	args := m.Called(messageID)
	return args.Bool(0)
}

// MockHook mocks a hook for testing
type MockHook struct {
	mock.Mock
}

func (m *MockHook) ShouldHandle(messageType string) bool {
	args := m.Called(messageType)
	return args.Bool(0)
}

func (m *MockHook) Validate(message *models.GossipMessage, msgType interfaces.MessageType) bool {
	args := m.Called(message, msgType)
	return args.Bool(0)
}

func (m *MockHook) Handle(message *models.GossipMessage, msgType interfaces.MessageType) error {
	args := m.Called(message, msgType)
	return args.Error(0)
}

func TestGossipProtocol_UpdatePeers(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "gossip_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.DefaultConfig(3000, 0)
	cfg.DataDir = tempDir

	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Create gossip protocol
	gossipProtocol := gossip.NewGossipProtocol(cfg, logger, mockStorage, mockHookManager)

	// Test UpdatePeers
	peers := []models.Peer{
		{
			NodeID:   "peer1",
			Address:  "127.0.0.1:3001",
			LastSeen: time.Now(),
		},
		{
			NodeID:   "peer2",
			Address:  "127.0.0.1:3002",
			LastSeen: time.Now(),
		},
	}

	gossipProtocol.UpdatePeers(peers)

	// Start the gossip protocol
	gossipProtocol.Start()
}

func TestGossipProtocol_HandleMessage(t *testing.T) {
	t.Run("TTL_Expired", func(t *testing.T) {
		mockStorage := new(MockStorage)
		mockHookManager := new(MockHookManager)
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		cfg := config.DefaultConfig(3000, 0)
		gossipProtocol := gossip.NewGossipProtocol(cfg, logger, mockStorage, mockHookManager)

		expiredMessage := &models.GossipMessage{
			MessageID:   "expired-message",
			OriginID:    "test-node",
			Timestamp:   time.Now().UTC(),
			TTL:         0,
			MessageType: "test_message",
			Payload:     map[string]interface{}{"content": "Expired message"},
		}

		err := gossipProtocol.HandleMessage(expiredMessage)
		if err != nil {
			t.Errorf("HandleMessage with expired TTL should succeed, got error: %v", err)
		}

		mockStorage.AssertExpectations(t)
		mockHookManager.AssertExpectations(t)
	})

	t.Run("Old_Message", func(t *testing.T) {
		mockStorage := new(MockStorage)
		mockHookManager := new(MockHookManager)
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		cfg := config.DefaultConfig(3000, 0)
		gossipProtocol := gossip.NewGossipProtocol(cfg, logger, mockStorage, mockHookManager)

		oldMessage := &models.GossipMessage{
			MessageID:   "old-message",
			OriginID:    "test-node",
			Timestamp:   time.Now().UTC().Add(-24 * time.Hour), // заведомо старое сообщение
			TTL:         5,
			MessageType: "test_message",
			Payload:     map[string]interface{}{"content": "Old message"},
		}

		err := gossipProtocol.HandleMessage(oldMessage)
		if err != nil {
			t.Errorf("HandleMessage with old timestamp should succeed, got error: %v", err)
		}

		mockStorage.AssertExpectations(t)
		mockHookManager.AssertExpectations(t)
	})

	t.Run("Already_Processed", func(t *testing.T) {
		mockStorage := new(MockStorage)
		mockHookManager := new(MockHookManager)
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		cfg := config.DefaultConfig(3000, 0)
		gossipProtocol := gossip.NewGossipProtocol(cfg, logger, mockStorage, mockHookManager)

		processedMessage := &models.GossipMessage{
			MessageID:   "processed-message",
			OriginID:    "test-node",
			Timestamp:   time.Now().UTC(),
			TTL:         5,
			MessageType: "test_message",
			Payload:     map[string]interface{}{"content": "Already processed"},
		}

		mockStorage.On("HasMessage", "processed-message").Return(true)

		err := gossipProtocol.HandleMessage(processedMessage)
		if err != nil {
			t.Errorf("HandleMessage with already processed message should succeed, got error: %v", err)
		}

		mockStorage.AssertExpectations(t)
		mockHookManager.AssertExpectations(t)
	})

	t.Run("Valid_Message", func(t *testing.T) {
		mockStorage := new(MockStorage)
		mockHookManager := new(MockHookManager)
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		cfg := config.DefaultConfig(3000, 0)
		gossipProtocol := gossip.NewGossipProtocol(cfg, logger, mockStorage, mockHookManager)

		validMessage := &models.GossipMessage{
			MessageID:   "valid-message",
			OriginID:    "test-node",
			Timestamp:   time.Now().UTC(),
			TTL:         5,
			MessageType: "test_message",
			Payload:     map[string]interface{}{"content": "Valid message"},
		}

		// Настраиваем моки для этого теста
		mockStorage.On("HasMessage", "valid-message").Return(false).Once() // первая проверка истории
		mockHookManager.On("ValidateMessage", mock.AnythingOfType("*models.GossipMessage"), interfaces.MessageTypePull).Return(true)
		mockHookManager.On("ProcessMessage", mock.AnythingOfType("*models.GossipMessage"), interfaces.MessageTypePull).Return(true)

		err := gossipProtocol.HandleMessage(validMessage)
		if err != nil {
			t.Errorf("HandleMessage with valid message failed: %v", err)
		}

		mockStorage.AssertExpectations(t)
		mockHookManager.AssertExpectations(t)
	})

	t.Run("Invalid_Message", func(t *testing.T) {
		mockStorage := new(MockStorage)
		mockHookManager := new(MockHookManager)
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		cfg := config.DefaultConfig(3000, 0)
		gossipProtocol := gossip.NewGossipProtocol(cfg, logger, mockStorage, mockHookManager)

		invalidMessage := &models.GossipMessage{
			MessageID:   "invalid-message",
			OriginID:    "test-node",
			Timestamp:   time.Now().UTC(),
			TTL:         5,
			MessageType: "test_message",
			Payload:     map[string]interface{}{"content": "Invalid message"},
		}

		// Настраиваем моки для этого теста
		mockStorage.On("HasMessage", "invalid-message").Return(false).Once()
		mockHookManager.On("ValidateMessage", mock.AnythingOfType("*models.GossipMessage"), interfaces.MessageTypePull).Return(false)

		err := gossipProtocol.HandleMessage(invalidMessage)
		if err == nil {
			t.Error("HandleMessage with invalid message should fail")
		} else if err.Error() != "message validation failed: invalid-message" {
			t.Errorf("Expected error 'message validation failed: invalid-message', got '%v'", err)
		}

		mockStorage.AssertExpectations(t)
		mockHookManager.AssertExpectations(t)
	})
}
