package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config содержит все настройки узла
type Config struct {
	NodeID        string         `json:"node_id"`
	Port          int            `json:"port"`
	DataDir       string         `json:"data_dir"`
	SeedNodes     []string       `json:"seed_nodes,omitempty"`
	GossipConfig  GossipConfig   `json:"gossip"`
	PexConfig     PexConfig      `json:"pex"`
	BlockchainConfig BlockchainConfig `json:"blockchain"`
}

// GossipConfig содержит настройки для Gossip протокола
type GossipConfig struct {
	ProtocolType     string        `json:"protocol_type"`
	BranchingFactor  int           `json:"branching_factor"`
	MessageTTL       int           `json:"message_ttl"`
	SyncInterval     time.Duration `json:"sync_interval"`
	HistorySize      int           `json:"history_size"`
	MessageMaxAge    time.Duration `json:"message_max_age"`
}

// PexConfig содержит настройки для PEX протокола
type PexConfig struct {
	ExchangeInterval   time.Duration `json:"exchange_interval"`
	MaxPeersPerExchange int          `json:"max_peers_per_exchange"`
	PeerTTL            time.Duration `json:"peer_ttl"`
	MaxPeers           int           `json:"max_peers"`
	NewPeerShare       int           `json:"new_peer_share"`
	LowConnectivityThreshold int     `json:"low_connectivity_threshold"`
}

// BlockchainConfig содержит настройки для блокчейна
type BlockchainConfig struct {
	BlockTime       time.Duration `json:"block_time"`
	MaxTransactions int           `json:"max_transactions"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig(port int, seedPort int) *Config {
	nodeID := fmt.Sprintf("node-%d", port)
	dataDir := filepath.Join(".nodedata", fmt.Sprintf("port%d", port))
	
	seedNodes := []string{}
	if seedPort > 0 && seedPort != port {
		seedNodes = append(seedNodes, fmt.Sprintf("127.0.0.1:%d", seedPort))
	}
	
	return &Config{
		NodeID:    nodeID,
		Port:      port,
		DataDir:   dataDir,
		SeedNodes: seedNodes,
		GossipConfig: GossipConfig{
			ProtocolType:    "push",
			BranchingFactor: 4,
			MessageTTL:      5,
			SyncInterval:    5 * time.Second,
			HistorySize:     10000,
			MessageMaxAge:   30 * time.Minute,
		},
		PexConfig: PexConfig{
			ExchangeInterval:         15 * time.Second,
			MaxPeersPerExchange:      10,
			PeerTTL:                  3 * time.Hour,
			MaxPeers:                 100,
			NewPeerShare:             5,
			LowConnectivityThreshold: 10,
		},
		BlockchainConfig: BlockchainConfig{
			BlockTime:       10 * time.Second,
			MaxTransactions: 100,
		},
	}
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}
	
	return &config, nil
}

// SaveConfig сохраняет конфигурацию в файл
func (c *Config) SaveConfig() error {
	// Убедимся, что директория существует
	configDir := filepath.Join(c.DataDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}
	
	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	
	return nil
}

// CreateDataDirs создает все необходимые директории для хранения данных
func (c *Config) CreateDataDirs() error {
	dirs := []string{
		filepath.Join(c.DataDir, "blocks"),
		filepath.Join(c.DataDir, "transactions"),
		filepath.Join(c.DataDir, "peers"),
		filepath.Join(c.DataDir, "config"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}

// CleanDataDirs удаляет все директории с данными
func (c *Config) CleanDataDirs() error {
	if err := os.RemoveAll(c.DataDir); err != nil {
		return fmt.Errorf("failed to clean data directory: %w", err)
	}
	return nil
}