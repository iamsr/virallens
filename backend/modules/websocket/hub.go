package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/iamsr/virallens/backend/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	UserIDs []uuid.UUID
	Message []byte
}

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
	go h.Run()
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			isFirstConnection := len(h.clients[client.UserID]) == 1
			h.mu.Unlock()
			log.Printf("Client connected: UserID=%s, ClientID=%s", client.UserID, client.ID)

			// Broadcast presence update only if it's their first connection
			if isFirstConnection {
				h.broadcastPresence(client.UserID.String(), "online")
			}

		case client := <-h.unregister:
			h.mu.Lock()
			isLastConnection := false
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
						isLastConnection = true
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: UserID=%s, ClientID=%s", client.UserID, client.ID)

			// Broadcast presence update only if it was their last connection
			if isLastConnection {
				h.broadcastPresence(client.UserID.String(), "offline")
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, userID := range message.UserIDs {
				if clients, ok := h.clients[userID]; ok {
					for client := range clients {
						select {
						case client.Send <- message.Message:
						default:
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

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToUsers(userIDs []uuid.UUID, message []byte) {
	h.broadcast <- &BroadcastMessage{
		UserIDs: userIDs,
		Message: message,
	}
}

func (h *Hub) BroadcastMessage(msg *models.Message, participants []uuid.UUID) error {
	wsMsg := WSMessage{
		Type: "message",
		Data: msg,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		return err
	}

	h.BroadcastToUsers(participants, data)
	return nil
}

func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// broadcastPresence sends a presence event to all currently connected clients.
func (h *Hub) broadcastPresence(userID string, status string) {
	presenceMsg := WSMessage{
		Type: "presence",
		Data: map[string]string{"user_id": userID, "status": status},
	}
	data, err := json.Marshal(presenceMsg)
	if err != nil {
		return
	}
	h.mu.RLock()
	var allClients []*Client
	for _, clients := range h.clients {
		for c := range clients {
			allClients = append(allClients, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range allClients {
		select {
		case c.Send <- data:
		default:
		}
	}
}

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

		if err := handler(c, message); err != nil {
			log.Printf("Error handling message: %v", err)
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
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

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

func (c *Client) StartPumps(handler func(*Client, []byte) error) {
	go c.writePump()
	go c.readPump(handler)
}

type WSMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type OutgoingMessage struct {
	Type           string  `json:"type"`
	ConversationID *string `json:"conversation_id,omitempty"`
	GroupID        *string `json:"group_id,omitempty"`
	Content        string  `json:"content"`
}
