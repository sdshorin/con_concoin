package models

import "time"

// GossipMessage представляет собой сообщение в Gossip протоколе
type GossipMessage struct {
	MessageID   string      `json:"message_id"`   // SHA-256 хеш
	OriginID    string      `json:"origin_id"`    // идентификатор_узла
	Timestamp   time.Time   `json:"timestamp"`    // UTC timestamp
	TTL         int         `json:"ttl"`          // число
	MessageType string      `json:"message_type"` // тип сообщения
	Payload     interface{} `json:"payload"`      // Содержимое сообщения
}

// PexMessage представляет собой сообщение в PEX протоколе
type PexMessage struct {
	MessageID string    `json:"message_id"` // уникальный_идентификатор
	Type      PexType   `json:"type"`       // pex_request|pex_response
	Timestamp time.Time `json:"timestamp"`  // UTC timestamp
	Peers     []Peer    `json:"peers"`      // Список пиров
}

// PexType тип сообщения в PEX протоколе
type PexType string

const (
	PexRequest  PexType = "pex_request"
	PexResponse PexType = "pex_response"
)

// Peer представляет собой информацию о пире
type Peer struct {
	NodeID   string    `json:"node_id"`   // идентификатор_узла
	Address  string    `json:"address"`   // IP:PORT
	LastSeen time.Time `json:"last_seen"` // UTC timestamp
}
