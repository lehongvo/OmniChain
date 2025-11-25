package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/onichange/pos-system/pkg/logger"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	UserID string
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	Logger *logger.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[string]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Redis client for pub/sub
	redis *redis.Client

	// Logger
	logger *logger.Logger

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	UserID    string      `json:"user_id,omitempty"`
	Channel   string      `json:"channel,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// NewHub creates a new WebSocket hub
func NewHub(redisClient *redis.Client, log *logger.Logger) *Hub {
	hub := &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      redisClient,
		logger:     log,
	}

	// Start Redis pub/sub listener
	go hub.listenRedis()

	return hub
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			h.logger.Infof("Client registered: %s (User: %s)", client.ID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			h.logger.Infof("Client unregistered: %s (User: %s)", client.ID, client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			// Broadcast to all clients
			for _, clients := range h.clients {
				for client := range clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}

// BroadcastToChannel publishes message to Redis channel for horizontal scaling
func (h *Hub) BroadcastToChannel(channel string, message []byte) error {
	return h.redis.Publish(context.Background(), channel, message).Err()
}

// listenRedis listens to Redis pub/sub for horizontal scaling
func (h *Hub) listenRedis() {
	pubsub := h.redis.Subscribe(context.Background(), "websocket:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		// Parse message and broadcast to local clients
		var wsMessage Message
		if err := json.Unmarshal([]byte(msg.Payload), &wsMessage); err != nil {
			h.logger.Errorf("Failed to unmarshal Redis message: %v", err)
			continue
		}

		// Broadcast to user if specified
		if wsMessage.UserID != "" {
			h.BroadcastToUser(wsMessage.UserID, []byte(msg.Payload))
		} else {
			// Broadcast to all
			h.broadcast <- []byte(msg.Payload)
		}
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.Logger.Errorf("Failed to unmarshal message: %v", err)
			continue
		}

		// Process message (e.g., subscribe to channel)
		if msg.Type == "subscribe" && msg.Channel != "" {
			// Subscribe to Redis channel
			c.Logger.Infof("Client %s subscribed to channel: %s", c.ID, msg.Channel)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID string, log *logger.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Logger: log,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the client's read and write pumps
func (c *Client) Start() {
	go c.WritePump()
	go c.ReadPump()
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	c.Conn.Close()
}
