package tests

import (
	"testing"
	"time"

	"concoin/conrun/pkg/hooks"
	"concoin/conrun/pkg/interfaces"
	"concoin/conrun/pkg/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// MockHook implements the Hook interface for testing
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

func TestHookManager(t *testing.T) {
	// Create logger for testing
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	// Create hook manager
	hookManager := hooks.NewHookManager("/tmp/test", logger)

	// Create mock hooks
	mockHook1 := new(MockHook)
	mockHook2 := new(MockHook)

	// Add hooks to manager
	hookManager.AddHook(mockHook1)
	hookManager.AddHook(mockHook2)

	// Test ValidateMessage with no matching hook
	mockHook1.On("ShouldHandle", "test_type").Return(false)
	mockHook2.On("ShouldHandle", "test_type").Return(false)

	message := &models.GossipMessage{
		MessageID:   "test-message-id",
		OriginID:    "test-node-id",
		Timestamp:   time.Now().UTC(),
		TTL:         5,
		MessageType: "test_type",
		Payload:     map[string]interface{}{"content": "Test content"},
	}

	result := hookManager.ValidateMessage(message, interfaces.MessageTypePull)
	if result {
		t.Errorf("ValidateMessage with no matching hook should return false")
	}

	// Reset mocks
	mockHook1.AssertExpectations(t)
	mockHook2.AssertExpectations(t)
	mockHook1 = new(MockHook)
	mockHook2 = new(MockHook)
	hookManager = hooks.NewHookManager("/tmp/test", logger)
	hookManager.AddHook(mockHook1)
	hookManager.AddHook(mockHook2)

	// Test ValidateMessage with matching hook that fails validation
	mockHook1.On("ShouldHandle", "test_type").Return(true)
	mockHook1.On("Validate", message, interfaces.MessageTypePull).Return(false)
	mockHook2.On("ShouldHandle", "test_type").Return(false)

	result = hookManager.ValidateMessage(message, interfaces.MessageTypePull)
	if result {
		t.Errorf("ValidateMessage with failing hook should return false")
	}

	// Reset mocks
	mockHook1.AssertExpectations(t)
	mockHook2.AssertExpectations(t)
	mockHook1 = new(MockHook)
	mockHook2 = new(MockHook)
	hookManager = hooks.NewHookManager("/tmp/test", logger)
	hookManager.AddHook(mockHook1)
	hookManager.AddHook(mockHook2)

	// Test ValidateMessage with matching hook that passes validation
	mockHook1.On("ShouldHandle", "test_type").Return(true)
	mockHook1.On("Validate", message, interfaces.MessageTypePull).Return(true)

	result = hookManager.ValidateMessage(message, interfaces.MessageTypePull)
	if !result {
		t.Errorf("ValidateMessage with passing hook should return true")
	}

	// Reset mocks
	mockHook1.AssertExpectations(t)
	mockHook1 = new(MockHook)
	mockHook2 = new(MockHook)
	hookManager = hooks.NewHookManager("/tmp/test", logger)
	hookManager.AddHook(mockHook1)
	hookManager.AddHook(mockHook2)

	// Test ProcessMessage with matching hook that passes validation and handles successfully
	mockHook1.On("ShouldHandle", "test_type").Return(true)
	mockHook1.On("Validate", message, interfaces.MessageTypePush).Return(true)
	mockHook1.On("Handle", message, interfaces.MessageTypePush).Return(nil)
	mockHook2.On("ShouldHandle", "test_type").Return(false)

	result = hookManager.ProcessMessage(message, interfaces.MessageTypePush)
	if !result {
		t.Errorf("ProcessMessage with passing hook should return true")
	}

	mockHook1.AssertExpectations(t)
	mockHook2.AssertExpectations(t)
}

func TestDebugHook(t *testing.T) {
	// Create logger for testing
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	// Create debug hook
	debugHook := hooks.NewDebugHook(logger)

	// Test ShouldHandle
	if !debugHook.ShouldHandle("user_message") {
		t.Errorf("Debug hook should handle 'user_message' type")
	}
	if debugHook.ShouldHandle("other_type") {
		t.Errorf("Debug hook should not handle 'other_type'")
	}

	// Test Validate
	message := &models.GossipMessage{
		MessageID:   "test-message-id",
		OriginID:    "test-node-id",
		Timestamp:   time.Now().UTC(),
		TTL:         5,
		MessageType: "user_message",
		Payload:     map[string]interface{}{"content": "Test content"},
	}

	if !debugHook.Validate(message, interfaces.MessageTypePull) {
		t.Errorf("Debug hook should validate all messages")
	}

	// Test Handle
	err := debugHook.Handle(message, interfaces.MessageTypePull)
	if err != nil {
		t.Errorf("Debug hook Handle failed: %v", err)
	}
}
