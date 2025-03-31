package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"concoin/conrun/pkg/api"
	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/interfaces"
	"concoin/conrun/pkg/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for dependencies
type MockGossipProtocol struct {
	mock.Mock
}

func (m *MockGossipProtocol) Start() {
	m.Called()
}

func (m *MockGossipProtocol) UpdatePeers(peers []models.Peer) {
	m.Called(peers)
}

func (m *MockGossipProtocol) HandleMessage(message *models.GossipMessage) error {
	args := m.Called(message)
	return args.Error(0)
}

type MockPexProtocol struct {
	mock.Mock
}

func (m *MockPexProtocol) Start() {
	m.Called()
}

func (m *MockPexProtocol) SetOnPeersListHandler(handler func(peers []models.Peer)) {
	m.Called(handler)
}

func (m *MockPexProtocol) AddPeer(peer models.Peer) bool {
	args := m.Called(peer)
	return args.Bool(0)
}

func (m *MockPexProtocol) GetPeers() []models.Peer {
	args := m.Called()
	return args.Get(0).([]models.Peer)
}

func (m *MockPexProtocol) HandlePexRequest(request models.PexMessage) models.PexMessage {
	args := m.Called(request)
	return args.Get(0).(models.PexMessage)
}

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

func TestAPI_handlePing(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Create test server
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	// Get the router from the API
	handler := nodeAPI.Router
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	if body := rr.Body.String(); body != "pong" {
		t.Errorf("Expected body 'pong', got %s", body)
	}
}

func TestAPI_handleGossipMessage(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Set up expectations
	testMessage := &models.GossipMessage{
		MessageID:   "test-message-id",
		OriginID:    "test-node-id",
		Timestamp:   time.Now().UTC(),
		TTL:         5,
		MessageType: "user_message",
		Payload:     map[string]interface{}{"content": "Hello, world!"},
	}

	mockGossip.On("HandleMessage", mock.AnythingOfType("*models.GossipMessage")).Return(nil)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Create test request
	messageData, err := json.Marshal(testMessage)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	req, err := http.NewRequest("POST", "/gossip", bytes.NewBuffer(messageData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Get the router from the API
	handler := nodeAPI.Router
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	mockGossip.AssertExpectations(t)
}

func TestAPI_handlePexMessage(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Set up expectations
	testPeer := models.Peer{
		NodeID:   "test-peer-id",
		Address:  "127.0.0.1:3001",
		LastSeen: time.Now().UTC(),
	}

	testRequest := models.PexMessage{
		MessageID: "test-pex-id",
		Type:      models.PexRequest,
		Timestamp: time.Now().UTC(),
		Peers:     []models.Peer{testPeer},
	}

	testResponse := models.PexMessage{
		MessageID: "test-response-id",
		Type:      models.PexResponse,
		Timestamp: time.Now().UTC(),
		Peers:     []models.Peer{testPeer},
	}

	mockPex.On("HandlePexRequest", mock.AnythingOfType("models.PexMessage")).Return(testResponse)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Create test request
	requestData, err := json.Marshal(testRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "/pex", bytes.NewBuffer(requestData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Get the router from the API
	handler := nodeAPI.Router
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Decode response
	var response models.PexMessage
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Type != models.PexResponse {
		t.Errorf("Expected response type PexResponse, got %s", response.Type)
	}

	mockPex.AssertExpectations(t)
}

func TestAPI_handleAddMessage(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Set up expectations
	mockHookManager.On("ProcessMessage", mock.AnythingOfType("*models.GossipMessage"), interfaces.MessageTypePush).Return(true)
	mockStorage.On("SaveMessage", mock.AnythingOfType("*models.GossipMessage")).Return(nil)
	mockGossip.On("HandleMessage", mock.AnythingOfType("*models.GossipMessage")).Return(nil)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Create test request
	requestData := []byte(`{
		"type": "user_message",
		"payload": {
			"content": "Hello, world!"
		}
	}`)

	req, err := http.NewRequest("POST", "/add_message", bytes.NewBuffer(requestData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Get the router from the API
	handler := nodeAPI.Router
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Decode response
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("Expected response status 'success', got %s", response["status"])
	}

	if response["message_id"] == "" {
		t.Errorf("Expected non-empty message_id in response")
	}

	mockHookManager.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockGossip.AssertExpectations(t)
}

func TestAPI_handleGetMessages(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Set up expectations
	mockStorage.On("GetMessageList").Return([]string{"msg1", "msg2"}, nil)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Create test request
	req, err := http.NewRequest("GET", "/messages", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	// Get the router from the API
	handler := nodeAPI.Router
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Decode response
	var messages []string
	if err := json.NewDecoder(rr.Body).Decode(&messages); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0] != "msg1" || messages[1] != "msg2" {
		t.Errorf("Expected messages [msg1, msg2], got %v", messages)
	}

	mockStorage.AssertExpectations(t)
}

func TestAPI_handleGetMessage(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Set up expectations
	testMessage := &models.GossipMessage{
		MessageID:   "test-message-id",
		OriginID:    "test-node-id",
		Timestamp:   time.Now().UTC(),
		TTL:         5,
		MessageType: "user_message",
		Payload:     map[string]interface{}{"content": "Hello, world!"},
	}

	mockStorage.On("GetMessage", "test-message-id").Return(testMessage, nil)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Create test request with path parameters
	req, err := http.NewRequest("GET", "/messages/test-message-id", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Создаем записывающий ответ
	rr := httptest.NewRecorder()

	// Устанавливаем переменные маршрута для теста
	// Создаем новый запрос с контекстом маршрутизатора, чтобы обходчик мог получить переменные
	req = mux.SetURLVars(req, map[string]string{
		"id": "test-message-id",
	})

	// Вместо использования маршрутизатора напрямую вызываем обработчик
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nodeAPI.Router.ServeHTTP(w, r)
	}).ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Decode response
	var message models.GossipMessage
	if err := json.NewDecoder(rr.Body).Decode(&message); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if message.MessageID != testMessage.MessageID {
		t.Errorf("Expected message ID %s, got %s", testMessage.MessageID, message.MessageID)
	}

	if message.OriginID != testMessage.OriginID {
		t.Errorf("Expected origin ID %s, got %s", testMessage.OriginID, message.OriginID)
	}

	if message.MessageType != testMessage.MessageType {
		t.Errorf("Expected message type %s, got %s", testMessage.MessageType, message.MessageType)
	}

	mockStorage.AssertExpectations(t)
}

func TestAPI_Start(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Start API in a goroutine with a random port (to avoid conflicts with real services)
	// This is just a basic test to ensure Start method doesn't crash
	go func() {
		nodeAPI.Start()
	}()

	// Wait a short time to make sure it has a chance to start
	time.Sleep(100 * time.Millisecond)
}

func TestAPI_LogHook(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	cfg := config.DefaultConfig(3000, 0)

	mockGossip := new(MockGossipProtocol)
	mockPex := new(MockPexProtocol)
	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Create API
	nodeAPI := api.NewAPI(cfg, mockGossip, mockPex, logger, mockStorage, mockHookManager)

	// Create log entry
	entry := &logrus.Entry{
		Message: "Test log message",
		Level:   logrus.InfoLevel,
	}

	// Send entry to hook
	nodeAPI.LogHook(entry)

	// No easy way to verify the log was added to the buffer without
	// exposing internal implementation details of the API.
	// This test just verifies that the LogHook method doesn't crash.
}
