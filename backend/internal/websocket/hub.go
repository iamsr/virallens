package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yourusername/virallens/backend/internal/domain"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a websocket client connection
type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub maintains active client connections and broadcasts messages
type Hub struct {
	// Registered clients mapped by user ID
	clients map[uuid.UUID]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to specific users
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage contains a message and target user IDs
type BroadcastMessage struct {
	UserIDs []uuid.UUID
	Message []byte
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("Client connected: UserID=%s, ClientID=%s", client.UserID, client.ID)

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
			log.Printf("Client disconnected: UserID=%s, ClientID=%s", client.UserID, client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, userID := range message.UserIDs {
				if clients, ok := h.clients[userID]; ok {
					for client := range clients {
						select {
						case client.Send <- message.Message:
						default:
							// Client's send buffer is full, close the connection
							close(client.Send)
							delete(clients, client)
							if len(clients) == 0 {
								delete(h.clients, userID)
							}
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient registers a new client with the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastToUsers sends a message to specific users
func (h *Hub) BroadcastToUsers(userIDs []uuid.UUID, message []byte) {
	h.broadcast <- &BroadcastMessage{
		UserIDs: userIDs,
		Message: message,
	}
}

// BroadcastMessage broadcasts a domain message to relevant users
func (h *Hub) BroadcastMessage(msg *domain.Message, participants []uuid.UUID) error {
	// Create WebSocket message
	wsMsg := WSMessage{
		Type: "message",
		Data: msg,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		return err
	}

	// Broadcast to all participants
	h.BroadcastToUsers(participants, data)
	return nil
}

// IsUserOnline checks if a user has any active connections
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetOnlineUsers returns a list of all online user IDs
func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump(handler func(*Client, []byte) error) {
	defer func() {
		c.Hub.UnregisterClient(c)
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
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle the message
		if err := handler(c, message); err != nil {
			log.Printf("Error handling message: %v", err)
			// Send error to client
			errMsg := WSMessage{
				Type:    "error",
				Message: err.Error(),
			}
			if data, err := json.Marshal(errMsg); err == nil {
				c.Send <- data
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
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
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
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

// StartPumps starts the read and write pumps for the client
func (c *Client) StartPumps(handler func(*Client, []byte) error) {
	go c.writePump()
	go c.readPump(handler)
}

// WSMessage represents a WebSocket message structure
type WSMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// OutgoingMessage represents a message sent by the client
type OutgoingMessage struct {
	Type           string     `json:"type"`
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	GroupID        *uuid.UUID `json:"group_id,omitempty"`
	Content        string     `json:"content"`
}
