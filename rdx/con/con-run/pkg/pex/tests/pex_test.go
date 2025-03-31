package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/interfaces"
	"concoin/conrun/pkg/models"
	"concoin/conrun/pkg/pex"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// These mock types are defined in the gossip tests but repeated here for completeness
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

func TestPexProtocol_AddPeer(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.ErrorLevel)

	// Create a test server for the ping check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Get server address
	serverAddr := server.Listener.Addr().String()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pex_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.MkdirAll(filepath.Join(tempDir, "peers"), 0755)

	cfg := config.DefaultConfig(3000, 0)
	cfg.DataDir = tempDir

	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Setup storage expectations
	mockStorage.On("SavePeer", mock.AnythingOfType("*models.Peer")).Return(nil)
	mockStorage.On("GetMessageList").Return([]string{}, nil)

	// Create pex protocol
	pexProtocol := pex.NewPexProtocol(cfg, mockStorage, logger, mockHookManager)

	// Test adding a peer
	peer := models.Peer{
		NodeID:   "test-peer",
		Address:  serverAddr,
		LastSeen: time.Now(),
	}

	result := pexProtocol.AddPeer(peer)
	if !result {
		t.Errorf("AddPeer failed to add valid peer")
	}

	// Test adding self address
	selfPeer := models.Peer{
		NodeID:   "self-peer",
		Address:  "127.0.0.1:3000",
		LastSeen: time.Now(),
	}

	result = pexProtocol.AddPeer(selfPeer)
	if result {
		t.Errorf("AddPeer should reject self address")
	}

	// Test adding invalid address
	invalidPeer := models.Peer{
		NodeID:   "invalid-peer",
		Address:  "invalid:address",
		LastSeen: time.Now(),
	}

	result = pexProtocol.AddPeer(invalidPeer)
	if result {
		t.Errorf("AddPeer should reject invalid address")
	}

	// Test adding unreachable peer
	unreachablePeer := models.Peer{
		NodeID:   "unreachable-peer",
		Address:  "127.0.0.1:54321", // Unlikely to have a server running here
		LastSeen: time.Now(),
	}

	result = pexProtocol.AddPeer(unreachablePeer)
	if result {
		t.Errorf("AddPeer should reject unreachable peer")
	}

	// Verify we can get peers
	peers := pexProtocol.GetPeers()
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(peers))
	}

	mockStorage.AssertExpectations(t)
}

func TestPexProtocol_HandlePexRequest(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.ErrorLevel)

	// Create a test server for the ping check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Get server address
	serverAddr := server.Listener.Addr().String()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pex_test_request")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.MkdirAll(filepath.Join(tempDir, "peers"), 0755)

	cfg := config.DefaultConfig(3000, 0)
	cfg.DataDir = tempDir

	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Setup storage expectations
	mockStorage.On("SavePeer", mock.AnythingOfType("*models.Peer")).Return(nil)
	mockStorage.On("GetMessageList").Return([]string{}, nil)

	// Create pex protocol
	pexProtocol := pex.NewPexProtocol(cfg, mockStorage, logger, mockHookManager)

	// Add a test peer to the protocol
	pexProtocol.AddPeer(models.Peer{
		NodeID:   "existing-peer",
		Address:  serverAddr,
		LastSeen: time.Now(),
	})

	// Create a test PEX request
	requestPeer := models.Peer{
		NodeID:   "request-peer",
		Address:  serverAddr,
		LastSeen: time.Now(),
	}

	request := models.PexMessage{
		MessageID: "test-request-id",
		Type:      models.PexRequest,
		Timestamp: time.Now().UTC(),
		Peers:     []models.Peer{requestPeer},
	}

	// Handle the request
	response := pexProtocol.HandlePexRequest(request)

	// Verify response
	if response.Type != models.PexResponse {
		t.Errorf("Expected response type PexResponse, got %s", response.Type)
	}

	if len(response.Peers) < 1 {
		t.Errorf("Expected at least 1 peer in response, got %d", len(response.Peers))
	}

	// Verify request peer was added
	peers := pexProtocol.GetPeers()
	found := false
	for _, p := range peers {
		if p.NodeID == requestPeer.NodeID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Request peer was not added to peer table")
	}

	mockStorage.AssertExpectations(t)
}

func TestPexProtocol_Start(t *testing.T) {
	// Create test dependencies
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.ErrorLevel)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pex_test_start")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.MkdirAll(filepath.Join(tempDir, "peers"), 0755)

	cfg := config.DefaultConfig(3000, 3001)
	cfg.DataDir = tempDir

	mockHookManager := new(MockHookManager)
	mockStorage := new(MockStorage)

	// Set up expectations for loading peers
	mockStorage.On("GetPeers").Return([]*models.Peer{}, nil)

	// Create pex protocol
	pexProtocol := pex.NewPexProtocol(cfg, mockStorage, logger, mockHookManager)

	// Set a test handler for peer list updates
	peerUpdateCalled := false
	pexProtocol.SetOnPeersListHandler(func(peers []models.Peer) {
		peerUpdateCalled = true
	})

	// Start the protocol
	pexProtocol.Start()

	// Wait a short time for seed nodes to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify seed nodes were added
	peers := pexProtocol.GetPeers()
	if len(peers) != 1 {
		t.Errorf("Expected 1 seed node, got %d peers", len(peers))
	}

	if !peerUpdateCalled {
		t.Errorf("Peer list update handler was not called")
	}

	mockStorage.AssertExpectations(t)
}
