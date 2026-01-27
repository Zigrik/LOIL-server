package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // В продакшене ограничить
	},
}

// ClientInfo - информация о клиенте
type ClientInfo struct {
	ID           string
	PlayerID     int
	CharacterID  int
	LocationID   int
	ConnectedAt  time.Time
	LastActivity time.Time
	IP           string
}

// Server - WebSocket сервер
type Server struct {
	Clients      map[string]*Client
	Broadcast    chan []byte
	Register     chan *Client
	Unregister   chan *Client
	Game         GameStateProvider
	UpdateTicker *time.Ticker
	mu           sync.RWMutex
	Sequence     int64
	Config       *ServerConfig
}

// Client - клиентское соединение (определение здесь, реализация в client.go)
type Client struct {
	Info     *ClientInfo
	Conn     *websocket.Conn
	Send     chan []byte
	Server   *Server
	mu       sync.Mutex
	sequence int64
}

// ServerConfig - конфигурация сервера
type ServerConfig struct {
	Addr           string
	UpdateInterval time.Duration
	PingInterval   time.Duration
	MaxMessageSize int64
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
}

// DefaultConfig - конфигурация по умолчанию
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		Addr:           ":8080",
		UpdateInterval: 100 * time.Millisecond,
		PingInterval:   30 * time.Second,
		MaxMessageSize: 1024 * 10,
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    60 * time.Second,
	}
}

// NewServer создает новый сервер
func NewServer(game GameStateProvider, config *ServerConfig) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	return &Server{
		Clients:      make(map[string]*Client),
		Broadcast:    make(chan []byte, 100),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Game:         game,
		UpdateTicker: time.NewTicker(config.UpdateInterval),
		Config:       config,
	}
}

// Start запускает сервер
func (s *Server) Start() error {
	// Запускаем обработчики
	go s.handleMessages()
	go s.broadcastUpdates()
	go s.sendPings()

	http.HandleFunc("/ws", s.serveWebSocket)
	http.HandleFunc("/health", s.healthCheck)

	log.Printf("Сервер запущен на %s", s.Config.Addr)
	log.Printf("Интервал обновлений: %v", s.Config.UpdateInterval)

	return http.ListenAndServe(s.Config.Addr, nil)
}

// serveWebSocket обрабатывает WebSocket соединения
func (s *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка WebSocket: %v", err)
		return
	}

	clientID := generateClientID()
	clientInfo := &ClientInfo{
		ID:           clientID,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		IP:           r.RemoteAddr,
	}

	client := &Client{
		Info:   clientInfo,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Server: s,
	}

	s.Register <- client

	// Запускаем обработчики клиента
	go client.writePump()
	go client.readPump()

	log.Printf("Клиент подключен: %s (%s)", clientID, r.RemoteAddr)
}

// healthCheck - проверка здоровья сервера
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	clientCount := len(s.Clients)
	s.mu.RUnlock()

	response := map[string]interface{}{
		"status":      "ok",
		"clients":     clientCount,
		"server_time": Now(),
		"update_rate": s.Config.UpdateInterval.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMessages обрабатывает системные сообщения
func (s *Server) handleMessages() {
	for {
		select {
		case client := <-s.Register:
			s.mu.Lock()
			s.Clients[client.Info.ID] = client
			s.mu.Unlock()

		case client := <-s.Unregister:
			s.mu.Lock()
			if _, ok := s.Clients[client.Info.ID]; ok {
				delete(s.Clients, client.Info.ID)
				close(client.Send)
				log.Printf("Клиент отключен: %s", client.Info.ID)
			}
			s.mu.Unlock()

		case message := <-s.Broadcast:
			s.mu.RLock()
			for _, client := range s.Clients {
				select {
				case client.Send <- message:
				default:
					log.Printf("Канал клиента %s переполнен", client.Info.ID)
					close(client.Send)
					delete(s.Clients, client.Info.ID)
				}
			}
			s.mu.RUnlock()
		}
	}
}

// broadcastUpdates рассылает обновления состояния
func (s *Server) broadcastUpdates() {
	for range s.UpdateTicker.C {
		s.mu.RLock()

		// Группируем клиентов по локациям
		clientsByLocation := make(map[int][]*Client)
		for _, client := range s.Clients {
			if client.Info.LocationID > 0 {
				clientsByLocation[client.Info.LocationID] = append(
					clientsByLocation[client.Info.LocationID], client)
			}
		}

		s.mu.RUnlock()

		// Рассылаем обновления для каждой локации
		for locationID, clients := range clientsByLocation {
			if len(clients) > 0 {
				update := s.createLocationUpdate(locationID)
				if update != nil {
					msg := Message{
						Type:    MsgLocationUpdate,
						Payload: update,
						Time:    Now(),
					}

					data, err := json.Marshal(msg)
					if err != nil {
						log.Printf("Ошибка маршалинга: %v", err)
						continue
					}

					// Отправляем всем клиентам в локации
					for _, client := range clients {
						client.sendRaw(data)
					}
				}
			}
		}
	}
}

// createLocationUpdate создает обновление локации
func (s *Server) createLocationUpdate(locationID int) *LocationUpdate {
	locationState := s.Game.GetLocationState(locationID)
	if locationState == nil {
		return nil
	}

	return &LocationUpdate{
		LocationID: locationID,
		Characters: s.Game.GetCharactersInLocation(locationID),
		Creatures:  s.Game.GetCreaturesInLocation(locationID),
		Objects:    s.Game.GetObjectsInLocation(locationID),
		ServerTime: s.Game.GetServerTime(),
	}
}

// sendPings отправляет ping сообщения
func (s *Server) sendPings() {
	ticker := time.NewTicker(s.Config.PingInterval)
	defer ticker.Stop()

	for range ticker.C {
		msg := Message{
			Type: MsgPing,
			Time: Now(),
		}

		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}

		s.mu.RLock()
		for _, client := range s.Clients {
			client.sendRaw(data)
		}
		s.mu.RUnlock()
	}
}

// sendRawToClient отправляет данные клиенту (вызов метода клиента)
func (s *Server) sendRawToClient(clientID string, data []byte) {
	s.mu.RLock()
	client, ok := s.Clients[clientID]
	s.mu.RUnlock()

	if ok {
		client.sendRaw(data)
	}
}

// Helper functions
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
