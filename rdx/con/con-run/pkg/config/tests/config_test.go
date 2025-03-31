package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"concoin/conrun/pkg/config"
)

func TestDefaultConfig(t *testing.T) {
	// Test creating a default config
	port := 3001
	seedPort := 3000
	cfg := config.DefaultConfig(port, seedPort)

	// Verify basic fields
	if cfg.Port != port {
		t.Errorf("Expected Port %d, got %d", port, cfg.Port)
	}
	if cfg.NodeID != "node-3001" {
		t.Errorf("Expected NodeID 'node-3001', got %s", cfg.NodeID)
	}
	if expected := ".nodedata/port3001"; cfg.DataDir != expected {
		t.Errorf("Expected DataDir %s, got %s", expected, cfg.DataDir)
	}

	// Verify seed nodes
	if len(cfg.SeedNodes) != 1 {
		t.Errorf("Expected 1 seed node, got %d", len(cfg.SeedNodes))
	} else if cfg.SeedNodes[0] != "127.0.0.1:3000" {
		t.Errorf("Expected seed node '127.0.0.1:3000', got %s", cfg.SeedNodes[0])
	}

	// Verify Gossip config
	if cfg.GossipConfig.ProtocolType != "push" {
		t.Errorf("Expected GossipConfig.ProtocolType 'push', got %s", cfg.GossipConfig.ProtocolType)
	}
	if cfg.GossipConfig.BranchingFactor != 4 {
		t.Errorf("Expected GossipConfig.BranchingFactor 4, got %d", cfg.GossipConfig.BranchingFactor)
	}
	if cfg.GossipConfig.MessageTTL != 5 {
		t.Errorf("Expected GossipConfig.MessageTTL 5, got %d", cfg.GossipConfig.MessageTTL)
	}
	if cfg.GossipConfig.SyncInterval != 5*time.Second {
		t.Errorf("Expected GossipConfig.SyncInterval 5s, got %v", cfg.GossipConfig.SyncInterval)
	}
	if cfg.GossipConfig.HistorySize != 10000 {
		t.Errorf("Expected GossipConfig.HistorySize 10000, got %d", cfg.GossipConfig.HistorySize)
	}
	if cfg.GossipConfig.MessageMaxAge != 30*time.Minute {
		t.Errorf("Expected GossipConfig.MessageMaxAge 30m, got %v", cfg.GossipConfig.MessageMaxAge)
	}

	// Verify PEX config
	if cfg.PexConfig.ExchangeInterval != 15*time.Second {
		t.Errorf("Expected PexConfig.ExchangeInterval 15s, got %v", cfg.PexConfig.ExchangeInterval)
	}
	if cfg.PexConfig.MaxPeersPerExchange != 10 {
		t.Errorf("Expected PexConfig.MaxPeersPerExchange 10, got %d", cfg.PexConfig.MaxPeersPerExchange)
	}
	if cfg.PexConfig.PeerTTL != 3*time.Hour {
		t.Errorf("Expected PexConfig.PeerTTL 3h, got %v", cfg.PexConfig.PeerTTL)
	}
	if cfg.PexConfig.MaxPeers != 100 {
		t.Errorf("Expected PexConfig.MaxPeers 100, got %d", cfg.PexConfig.MaxPeers)
	}
	if cfg.PexConfig.NewPeerShare != 5 {
		t.Errorf("Expected PexConfig.NewPeerShare 5, got %d", cfg.PexConfig.NewPeerShare)
	}
	if cfg.PexConfig.LowConnectivityThreshold != 10 {
		t.Errorf("Expected PexConfig.LowConnectivityThreshold 10, got %d", cfg.PexConfig.LowConnectivityThreshold)
	}

	// Verify Blockchain config
	if cfg.BlockchainConfig.BlockTime != 10*time.Second {
		t.Errorf("Expected BlockchainConfig.BlockTime 10s, got %v", cfg.BlockchainConfig.BlockTime)
	}
	if cfg.BlockchainConfig.MaxTransactions != 100 {
		t.Errorf("Expected BlockchainConfig.MaxTransactions 100, got %d", cfg.BlockchainConfig.MaxTransactions)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config
	cfg := config.DefaultConfig(3001, 3000)
	cfg.DataDir = tempDir

	// Save the config
	err = cfg.SaveConfig()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify the config file exists
	configPath := filepath.Join(tempDir, "config", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file does not exist: %s", configPath)
	}

	// Load the config
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config matches original
	if loadedCfg.NodeID != cfg.NodeID {
		t.Errorf("Loaded NodeID %s doesn't match original %s", loadedCfg.NodeID, cfg.NodeID)
	}
	if loadedCfg.Port != cfg.Port {
		t.Errorf("Loaded Port %d doesn't match original %d", loadedCfg.Port, cfg.Port)
	}
	if loadedCfg.DataDir != cfg.DataDir {
		t.Errorf("Loaded DataDir %s doesn't match original %s", loadedCfg.DataDir, cfg.DataDir)
	}
	if len(loadedCfg.SeedNodes) != len(cfg.SeedNodes) {
		t.Errorf("Loaded SeedNodes length %d doesn't match original %d", len(loadedCfg.SeedNodes), len(cfg.SeedNodes))
	} else if loadedCfg.SeedNodes[0] != cfg.SeedNodes[0] {
		t.Errorf("Loaded SeedNode %s doesn't match original %s", loadedCfg.SeedNodes[0], cfg.SeedNodes[0])
	}

	// Verify Gossip config
	if loadedCfg.GossipConfig.ProtocolType != cfg.GossipConfig.ProtocolType {
		t.Errorf("Loaded GossipConfig.ProtocolType %s doesn't match original %s", 
			loadedCfg.GossipConfig.ProtocolType, cfg.GossipConfig.ProtocolType)
	}
	if loadedCfg.GossipConfig.MessageTTL != cfg.GossipConfig.MessageTTL {
		t.Errorf("Loaded GossipConfig.MessageTTL %d doesn't match original %d", 
			loadedCfg.GossipConfig.MessageTTL, cfg.GossipConfig.MessageTTL)
	}
}

func TestCreateDataDirs(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "config_test_dirs")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config
	cfg := config.DefaultConfig(3001, 3000)
	cfg.DataDir = tempDir

	// Create data directories
	err = cfg.CreateDataDirs()
	if err != nil {
		t.Fatalf("Failed to create data dirs: %v", err)
	}

	// Verify directories exist
	dirs := []string{
		filepath.Join(tempDir, "blocks"),
		filepath.Join(tempDir, "transactions"),
		filepath.Join(tempDir, "peers"),
		filepath.Join(tempDir, "config"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory does not exist: %s", dir)
		}
	}
}

func TestCleanDataDirs(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "config_test_clean")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a test config
	cfg := config.DefaultConfig(3001, 3000)
	cfg.DataDir = tempDir

	// Create data directories
	err = cfg.CreateDataDirs()
	if err != nil {
		t.Fatalf("Failed to create data dirs: %v", err)
	}

	// Clean data directories
	err = cfg.CleanDataDirs()
	if err != nil {
		t.Fatalf("Failed to clean data dirs: %v", err)
	}

	// Verify directory no longer exists
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("Directory still exists after cleanup: %s", tempDir)
	}
}