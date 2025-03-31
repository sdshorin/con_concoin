package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"concoin/conrun/pkg/models"
	"concoin/conrun/pkg/storage"
)

func TestStorage(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store := storage.NewStorage(tempDir)

	// Test peer storage
	t.Run("TestPeerStorage", func(t *testing.T) {
		// Create test peer
		peer := &models.Peer{
			NodeID:   "test-peer-id",
			Address:  "127.0.0.1:3000",
			LastSeen: time.Now().UTC(),
		}

		// Save peer
		err := store.SavePeer(peer)
		if err != nil {
			t.Fatalf("Failed to save peer: %v", err)
		}

		// Verify peer file exists
		peerPath := filepath.Join(tempDir, "peers", "test-peer-id.json")
		if _, err := os.Stat(peerPath); os.IsNotExist(err) {
			t.Fatalf("Peer file does not exist: %s", peerPath)
		}

		// Get peers
		peers, err := store.GetPeers()
		if err != nil {
			t.Fatalf("Failed to get peers: %v", err)
		}

		// Verify peer is returned
		if len(peers) != 1 {
			t.Fatalf("Expected 1 peer, got %d", len(peers))
		}
		if peers[0].NodeID != peer.NodeID {
			t.Errorf("Expected peer NodeID %s, got %s", peer.NodeID, peers[0].NodeID)
		}
		if peers[0].Address != peer.Address {
			t.Errorf("Expected peer Address %s, got %s", peer.Address, peers[0].Address)
		}
	})

	// Test message storage
	t.Run("TestMessageStorage", func(t *testing.T) {
		// Create test message
		message := &models.GossipMessage{
			MessageID:   "test-message-id",
			OriginID:    "test-node-id",
			Timestamp:   time.Now().UTC(),
			TTL:         5,
			MessageType: "user_message",
			Payload:     map[string]interface{}{"content": "Hello, world!"},
		}

		// Save message
		err := store.SaveMessage(message)
		if err != nil {
			t.Fatalf("Failed to save message: %v", err)
		}

		// Verify message file exists
		messagePath := filepath.Join(tempDir, "messages", "test-message-id.json")
		if _, err := os.Stat(messagePath); os.IsNotExist(err) {
			t.Fatalf("Message file does not exist: %s", messagePath)
		}

		// Test HasMessage
		if !store.HasMessage("test-message-id") {
			t.Errorf("HasMessage returned false for existing message")
		}
		if store.HasMessage("non-existent-id") {
			t.Errorf("HasMessage returned true for non-existent message")
		}

		// Get message
		retrievedMsg, err := store.GetMessage("test-message-id")
		if err != nil {
			t.Fatalf("Failed to get message: %v", err)
		}

		// Verify message fields
		if retrievedMsg.MessageID != message.MessageID {
			t.Errorf("Expected MessageID %s, got %s", message.MessageID, retrievedMsg.MessageID)
		}
		if retrievedMsg.OriginID != message.OriginID {
			t.Errorf("Expected OriginID %s, got %s", message.OriginID, retrievedMsg.OriginID)
		}
		if retrievedMsg.TTL != message.TTL {
			t.Errorf("Expected TTL %d, got %d", message.TTL, retrievedMsg.TTL)
		}
		if retrievedMsg.MessageType != message.MessageType {
			t.Errorf("Expected MessageType %s, got %s", message.MessageType, retrievedMsg.MessageType)
		}

		// Get message list
		messageIDs, err := store.GetMessageList()
		if err != nil {
			t.Fatalf("Failed to get message list: %v", err)
		}

		// Verify message list
		if len(messageIDs) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messageIDs))
		}
		if messageIDs[0] != message.MessageID {
			t.Errorf("Expected message ID %s, got %s", message.MessageID, messageIDs[0])
		}
	})

}
