package tests

import (
	"encoding/json"
	"testing"
	"time"

	"concoin/conrun/pkg/models"
)

func TestGossipMessageSerialization(t *testing.T) {
	// Create a test message
	message := models.GossipMessage{
		MessageID:   "test-message-id",
		OriginID:    "test-node-id",
		Timestamp:   time.Now().UTC(),
		TTL:         5,
		MessageType: "user_message",
		Payload:     map[string]interface{}{"content": "Hello, world!"},
	}

	// Serialize the message
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// Deserialize the message
	var unmarshaled models.GossipMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	// Verify fields
	if unmarshaled.MessageID != message.MessageID {
		t.Errorf("Expected MessageID %s, got %s", message.MessageID, unmarshaled.MessageID)
	}
	if unmarshaled.OriginID != message.OriginID {
		t.Errorf("Expected OriginID %s, got %s", message.OriginID, unmarshaled.OriginID)
	}
	if unmarshaled.TTL != message.TTL {
		t.Errorf("Expected TTL %d, got %d", message.TTL, unmarshaled.TTL)
	}
	if unmarshaled.MessageType != message.MessageType {
		t.Errorf("Expected MessageType %s, got %s", message.MessageType, unmarshaled.MessageType)
	}

	// Check timestamp (approximate equality due to serialization precision)
	timeDiff := unmarshaled.Timestamp.Sub(message.Timestamp)
	if timeDiff < -time.Second || timeDiff > time.Second {
		t.Errorf("Timestamp differs too much: original %v, unmarshaled %v", message.Timestamp, unmarshaled.Timestamp)
	}

	// Check payload
	payloadMap, ok := unmarshaled.Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected payload to be a map, got %T", unmarshaled.Payload)
	}
	content, ok := payloadMap["content"]
	if !ok {
		t.Fatalf("Expected payload to have content field")
	}
	if content != "Hello, world!" {
		t.Errorf("Expected content to be 'Hello, world!', got %v", content)
	}
}

func TestPexMessageSerialization(t *testing.T) {
	// Create a test peer
	peer := models.Peer{
		NodeID:   "test-peer-id",
		Address:  "127.0.0.1:3000",
		LastSeen: time.Now().UTC(),
	}

	// Create a test PEX message
	message := models.PexMessage{
		MessageID: "test-pex-id",
		Type:      models.PexRequest,
		Timestamp: time.Now().UTC(),
		Peers:     []models.Peer{peer},
	}

	// Serialize the message
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal PEX message: %v", err)
	}

	// Deserialize the message
	var unmarshaled models.PexMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal PEX message: %v", err)
	}

	// Verify fields
	if unmarshaled.MessageID != message.MessageID {
		t.Errorf("Expected MessageID %s, got %s", message.MessageID, unmarshaled.MessageID)
	}
	if unmarshaled.Type != message.Type {
		t.Errorf("Expected Type %s, got %s", message.Type, unmarshaled.Type)
	}

	// Check timestamp
	timeDiff := unmarshaled.Timestamp.Sub(message.Timestamp)
	if timeDiff < -time.Second || timeDiff > time.Second {
		t.Errorf("Timestamp differs too much: original %v, unmarshaled %v", message.Timestamp, unmarshaled.Timestamp)
	}

	// Check peers
	if len(unmarshaled.Peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(unmarshaled.Peers))
	}
	
	unmarshaledPeer := unmarshaled.Peers[0]
	if unmarshaledPeer.NodeID != peer.NodeID {
		t.Errorf("Expected peer NodeID %s, got %s", peer.NodeID, unmarshaledPeer.NodeID)
	}
	if unmarshaledPeer.Address != peer.Address {
		t.Errorf("Expected peer Address %s, got %s", peer.Address, unmarshaledPeer.Address)
	}
	
	// Check peer timestamp
	peerTimeDiff := unmarshaledPeer.LastSeen.Sub(peer.LastSeen)
	if peerTimeDiff < -time.Second || peerTimeDiff > time.Second {
		t.Errorf("Peer LastSeen differs too much: original %v, unmarshaled %v", peer.LastSeen, unmarshaledPeer.LastSeen)
	}
}

func TestPexTypeConstants(t *testing.T) {
	// Verify the PEX type constants
	if models.PexRequest != "pex_request" {
		t.Errorf("Expected PexRequest to be 'pex_request', got %s", models.PexRequest)
	}
	if models.PexResponse != "pex_response" {
		t.Errorf("Expected PexResponse to be 'pex_response', got %s", models.PexResponse)
	}
}