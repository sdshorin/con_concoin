package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"concoin/conrun/pkg/config"
	"concoin/conrun/pkg/interfaces"
	"concoin/conrun/pkg/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// API представляет собой HTTP API узла
type API struct {
	config      *config.Config
	gossip      interfaces.GossipProtocolInterface
	pex         interfaces.PexProtocolInterface
	logger      *logrus.Logger
	logBuffer   []LogEntry
	Router      *mux.Router
	storage     interfaces.StorageInterface
	hookManager interfaces.HookManagerInterface
}

// LogEntry представляет собой запись лога
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// NodeStats представляет собой статистику узла
type NodeStats struct {
	NodeID  string `json:"node_id"`
	Address string `json:"address"`
	Peers   int    `json:"peers"`
	Uptime  string `json:"uptime"`
}

// NetworkStats представляет собой статистику сети
type NetworkStats struct {
	LocalNode NodeStats   `json:"local_node"`
	Peers     []PeerStats `json:"peers"`
}

// PeerStats представляет собой статистику пира
type PeerStats struct {
	NodeID   string    `json:"node_id"`
	Address  string    `json:"address"`
	LastSeen time.Time `json:"last_seen"`
	IsOnline bool      `json:"is_online"`
}

// NewAPI создает новый экземпляр API
func NewAPI(config *config.Config, gossip interfaces.GossipProtocolInterface, pex interfaces.PexProtocolInterface, logger *logrus.Logger, storage interfaces.StorageInterface, hookManager interfaces.HookManagerInterface) *API {
	api := &API{
		config:      config,
		gossip:      gossip,
		pex:         pex,
		logger:      logger,
		logBuffer:   make([]LogEntry, 0, 100),
		Router:      mux.NewRouter(),
		storage:     storage,
		hookManager: hookManager,
	}

	// Настраиваем маршруты
	api.setupRoutes()

	return api
}

// setupRoutes настраивает маршруты API
func (a *API) setupRoutes() {
	// Gossip и PEX обработчики
	a.Router.HandleFunc("/gossip", a.handleGossipMessage).Methods("POST")
	a.Router.HandleFunc("/pex", a.handlePexMessage).Methods("POST")

	// Проверка доступности
	a.Router.HandleFunc("/ping", a.handlePing).Methods("GET")

	// Отладочный API
	a.Router.HandleFunc("/debug", a.handleDebug).Methods("GET")
	a.Router.HandleFunc("/network", a.handleNetwork).Methods("GET")

	// API для работы с сообщениями
	a.Router.HandleFunc("/messages", a.handleGetMessages).Methods("GET")
	a.Router.HandleFunc("/messages/{id}", a.handleGetMessage).Methods("GET")
	a.Router.HandleFunc("/message", a.handleMessage).Methods("POST")
	a.Router.HandleFunc("/add_message", a.handleAddMessage).Methods("POST")
}

// Start запускает HTTP сервер
func (a *API) Start() {
	// Запускаем сервер на указанном порту
	addr := fmt.Sprintf(":%d", a.config.Port)
	go func() {
		a.logger.Infof("Starting API server on %s", addr)
		if err := http.ListenAndServe(addr, a.Router); err != nil {
			a.logger.Fatalf("Failed to start API server: %v", err)
		}
	}()
}

// LogHook представляет собой хук для logrus
func (a *API) LogHook(entry *logrus.Entry) {
	// Добавляем запись в буфер логов
	logEntry := LogEntry{
		Timestamp: time.Now(),
		Level:     entry.Level.String(),
		Message:   entry.Message,
	}

	// Ограничиваем размер буфера
	if len(a.logBuffer) >= 100 {
		a.logBuffer = a.logBuffer[1:]
	}

	a.logBuffer = append(a.logBuffer, logEntry)
}

// handleGossipMessage обрабатывает входящее Gossip сообщение
func (a *API) handleGossipMessage(w http.ResponseWriter, r *http.Request) {
	var message models.GossipMessage

	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		a.logger.Warnf("Failed to decode Gossip message: %v", err)
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Обрабатываем сообщение
	if err := a.gossip.HandleMessage(&message); err != nil {
		a.logger.Warnf("Failed to handle Gossip message: %v", err)
		http.Error(w, "Failed to process message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handlePexMessage обрабатывает входящее PEX сообщение
func (a *API) handlePexMessage(w http.ResponseWriter, r *http.Request) {
	var request models.PexMessage

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		a.logger.Warnf("Failed to decode PEX message: %v", err)
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Проверяем тип сообщения
	if request.Type != models.PexRequest {
		http.Error(w, "Invalid message type", http.StatusBadRequest)
		return
	}

	// Обрабатываем запрос и формируем ответ
	response := a.pex.HandlePexRequest(request)

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.logger.Warnf("Failed to encode PEX response: %v", err)
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}
}

// handlePing обрабатывает запрос проверки доступности
func (a *API) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// handleMessage обрабатывает общий запрос для передачи данных
func (a *API) handleMessage(w http.ResponseWriter, r *http.Request) {
	var message models.GossipMessage

	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		a.logger.Warnf("Failed to decode message: %v", err)
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Обрабатываем сообщение через хуки
	isValid := a.hookManager.ProcessMessage(&message, interfaces.MessageTypePush)

	if !isValid {
		a.logger.Warnf("Message validation failed: %s", message.MessageID)
		http.Error(w, "Message validation failed", http.StatusBadRequest)
		return
	}

	// Сохраняем сообщение
	if err := a.storage.SaveMessage(&message); err != nil {
		a.logger.Warnf("Failed to save message: %v", err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	// Отправляем сообщение через gossip
	if err := a.gossip.HandleMessage(&message); err != nil {
		a.logger.Warnf("Failed to propagate message via gossip: %v", err)
		// Не возвращаем ошибку, т.к. сообщение уже сохранено
	}

	w.WriteHeader(http.StatusOK)
}

// handleDebug обрабатывает запрос отладочной информации
func (a *API) handleDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Получаем статистику узла
	stats := a.getNodeStats()

	// Формируем HTML-страницу
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Node Debug - {{.NodeID}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2 { color: #333; }
        .stats { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .logs { background: #f5f5f5; padding: 15px; border-radius: 5px; height: 400px; overflow-y: scroll; }
        table { width: 100%; border-collapse: collapse; }
        th, td { text-align: left; padding: 8px; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        .error { color: #d32f2f; }
        .warn { color: #ff9800; }
        .info { color: #2196f3; }
        .debug { color: #4caf50; }
        .messages { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-top: 20px; }
        pre { margin: 0; white-space: pre-wrap; word-wrap: break-word; max-width: 500px; }
        .message-cell { max-width: 500px; }
        .nav { margin-bottom: 20px; }
        .nav a { 
            display: inline-block;
            padding: 8px 16px;
            background-color: #2196f3;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin-right: 10px;
        }
        .nav a:hover {
            background-color: #1976d2;
        }
    </style>
</head>
<body>
    <div class="nav">
        <a href="/debug">Debug</a>
        <a href="/network">Network</a>
    </div>
    <h1>Node Debug - {{.NodeID}}</h1>
    
    <h2>Node Statistics</h2>
    <div class="stats">
        <table>
            <tr>
                <th>Node ID</th>
                <td>{{.NodeID}}</td>
            </tr>
            <tr>
                <th>Address</th>
                <td>{{.Address}}</td>
            </tr>
            <tr>
                <th>Connected Peers</th>
                <td>{{.Peers}}</td>
            </tr>
            <tr>
                <th>Node Start Time</th>
                <td>{{.Uptime}}</td>
            </tr>
        </table>
    </div>
    
    <h2>Node Logs</h2>
    <div class="logs">
        <table>
            <tr>
                <th>Time</th>
                <th>Level</th>
                <th>Message</th>
            </tr>
            {{range .Logs}}
            <tr class="{{.Level}}">
                <td>{{.Timestamp.Format "15:04:05"}}</td>
                <td>{{.Level}}</td>
                <td>{{.Message}}</td>
            </tr>
            {{end}}
        </table>
    </div>

    <h2>Messages</h2>
    <div class="messages">
        <table>
            <tr>
                <th>ID</th>
                <th>Type</th>
                <th>Origin</th>
                <th>Time</th>
                <th>TTL</th>
                <th>Payload</th>
            </tr>
            {{range .Messages}}
            <tr>
                <td>{{.MessageID}}</td>
                <td>{{.MessageType}}</td>
                <td>{{.OriginID}}</td>
                <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
                <td>{{.TTL}}</td>
                <td class="message-cell"><pre>{{.Payload}}</pre></td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>
`

	// Получаем список сообщений
	messageIDs, err := a.storage.GetMessageList()
	if err != nil {
		a.logger.Warnf("Failed to get message list: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	// Получаем все сообщения
	var messages []models.GossipMessage
	for _, msgID := range messageIDs {
		msg, err := a.storage.GetMessage(msgID)
		if err != nil {
			a.logger.Warnf("Failed to get message %s: %v", msgID, err)
			continue
		}
		messages = append(messages, *msg)
	}

	// Создаем данные для шаблона
	data := struct {
		NodeID   string
		Address  string
		Peers    int
		Uptime   string
		Logs     []LogEntry
		Messages []models.GossipMessage
	}{
		NodeID:   stats.NodeID,
		Address:  stats.Address,
		Peers:    stats.Peers,
		Uptime:   stats.Uptime,
		Logs:     a.logBuffer,
		Messages: messages,
	}

	// Парсим и выполняем шаблон
	t, err := template.New("debug").Parse(tmpl)
	if err != nil {
		a.logger.Warnf("Failed to parse template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		a.logger.Warnf("Failed to execute template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleNetwork обрабатывает запрос информации о сети
func (a *API) handleNetwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Получаем статистику узла и пиров
	localStats := a.getNodeStats()
	peersList := a.pex.GetPeers()

	peerStats := make([]PeerStats, 0, len(peersList))
	for _, peer := range peersList {
		// Проверяем доступность пира
		isOnline := a.checkPeerStatus(peer.Address)

		peerStats = append(peerStats, PeerStats{
			NodeID:   peer.NodeID,
			Address:  peer.Address,
			LastSeen: peer.LastSeen,
			IsOnline: isOnline,
		})
	}

	// Формируем HTML-страницу
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Network Stats - {{.LocalNode.NodeID}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2 { color: #333; }
        .stats { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { text-align: left; padding: 8px; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        .online { color: green; }
        .offline { color: red; }
        .nav { margin-bottom: 20px; }
        .nav a { 
            display: inline-block;
            padding: 8px 16px;
            background-color: #2196f3;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin-right: 10px;
        }
        .nav a:hover {
            background-color: #1976d2;
        }
    </style>
</head>
<body>
    <div class="nav">
        <a href="/debug">Debug</a>
        <a href="/network">Network</a>
    </div>
    <h1>Network Stats - {{.LocalNode.NodeID}}</h1>
    
    <h2>Local Node</h2>
    <div class="stats">
        <table>
            <tr>
                <th>Node ID</th>
                <td>{{.LocalNode.NodeID}}</td>
            </tr>
            <tr>
                <th>Address</th>
                <td>{{.LocalNode.Address}}</td>
            </tr>
            <tr>
                <th>Connected Peers</th>
                <td>{{.LocalNode.Peers}}</td>
            </tr>
            <tr>
                <th>Node Start Time</th>
                <td>{{.LocalNode.Uptime}}</td>
            </tr>
        </table>
    </div>
    
    <h2>Connected Peers</h2>
    <div class="stats">
        <table>
            <tr>
                <th>Node ID</th>
                <th>Address</th>
                <th>Last Seen</th>
                <th>Status</th>
                <th>Action</th>
            </tr>
            {{range .Peers}}
            <tr>
                <td>{{.NodeID}}</td>
                <td>{{.Address}}</td>
                <td>{{.LastSeen.Format "2006-01-02 15:04:05"}}</td>
                <td class="{{if .IsOnline}}online{{else}}offline{{end}}">
                    {{if .IsOnline}}Online{{else}}Offline{{end}}
                </td>
                <td>
                    <a href="http://{{.Address}}/debug" target="_blank">Debug</a>
                </td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>
`

	// Создаем данные для шаблона
	data := struct {
		LocalNode NodeStats
		Peers     []PeerStats
	}{
		LocalNode: localStats,
		Peers:     peerStats,
	}

	// Парсим и выполняем шаблон
	t, err := template.New("network").Parse(tmpl)
	if err != nil {
		a.logger.Warnf("Failed to parse template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		a.logger.Warnf("Failed to execute template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// getNodeStats возвращает статистику узла
func (a *API) getNodeStats() NodeStats {
	peers := a.pex.GetPeers()

	return NodeStats{
		NodeID:  a.config.NodeID,
		Address: fmt.Sprintf("127.0.0.1:%d", a.config.Port),
		Peers:   len(peers),
		Uptime:  "N/A", // В MVP не отслеживаем время работы
	}
}

// checkPeerStatus проверяет доступность пира
func (a *API) checkPeerStatus(address string) bool {
	url := fmt.Sprintf("http://%s/ping", address)
	client := http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// handleGetMessages обрабатывает запрос списка всех сообщений
func (a *API) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := a.storage.GetMessageList()
	if err != nil {
		a.logger.Warnf("Failed to get message list: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// handleGetMessage обрабатывает запрос конкретного сообщения
func (a *API) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	message, err := a.storage.GetMessage(messageID)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Message not found", http.StatusNotFound)
			return
		}
		a.logger.Warnf("Failed to get message %s: %v", messageID, err)
		http.Error(w, "Failed to get message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

// handleAddMessage обрабатывает запрос на добавление нового сообщения
func (a *API) handleAddMessage(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	var request struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}

	projectRoot, err := os.Getwd()
	if err != nil {
	}
	a.logger.Infof("handleAddMessage - Project root directory: %s", projectRoot)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		a.logger.Warnf("Failed to decode message payload: %v", err)
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Создаем Gossip сообщение
	message := &models.GossipMessage{
		MessageID:   fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		OriginID:    a.config.NodeID,
		Timestamp:   time.Now().UTC(),
		TTL:         a.config.GossipConfig.MessageTTL,
		MessageType: request.Type,
		Payload:     request.Payload,
	}

	// Обрабатываем сообщение через хуки
	isValid := a.hookManager.ProcessMessage(message, interfaces.MessageTypePush)

	if !isValid {
		a.logger.Warnf("Message validation failed: %s", message.MessageID)
		http.Error(w, "Message validation failed", http.StatusBadRequest)
		return
	}

	// Сохраняем сообщение
	if err := a.storage.SaveMessage(message); err != nil {
		a.logger.Warnf("Failed to save message: %v", err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	// Отправляем сообщение через gossip
	if err := a.gossip.HandleMessage(message); err != nil {
		a.logger.Warnf("Failed to propagate message via gossip: %v", err)
		// Не возвращаем ошибку, т.к. сообщение уже сохранено
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"message_id": message.MessageID,
	})
}
