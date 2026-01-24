package chat

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev; restrict in production
	},
}

// Client represents a WebSocket client connection
type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	userID         string
	conversationID string
}

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	// Registered clients by conversation ID
	conversations map[string]map[*Client]bool
	// Register requests from clients
	register chan *Client
	// Unregister requests from clients
	unregister chan *Client
	// Inbound messages from clients
	broadcast chan *BroadcastMessage
	// Mutex for thread-safe access
	mu sync.RWMutex
}

// BroadcastMessage wraps a message with its target conversation
type BroadcastMessage struct {
	ConversationID string
	Message        []byte
	SenderID       string
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		conversations: make(map[string]map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan *BroadcastMessage),
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.conversations[client.conversationID] == nil {
				h.conversations[client.conversationID] = make(map[*Client]bool)
			}
			h.conversations[client.conversationID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.conversations[client.conversationID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.conversations, client.conversationID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.conversations[message.ConversationID]
			h.mu.RUnlock()

			data := message.Message

			for client := range clients {
				// Don't send to the sender
				// Don't send to the sender
				if client.userID == message.SenderID {
					continue
				}
				select {
				case client.send <- data:
				default:
					h.mu.Lock()
					delete(h.conversations[message.ConversationID], client)
					close(client.send)
					h.mu.Unlock()
				}
			}
		}
	}
}

// Broadcast sends a message to all clients in a conversation
func (h *Hub) Broadcast(conversationID string, msg []byte, senderID string) {
	h.broadcast <- &BroadcastMessage{
		ConversationID: conversationID,
		Message:        msg,
		SenderID:       senderID,
	}
}

// HandleWebSocket handles WebSocket upgrade and client management
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID := r.URL.Query().Get("conversation")
	if conversationID == "" {
		http.Error(w, "Missing conversation ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:            h,
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         sessionData.UserID,
		conversationID: conversationID,
	}
	h.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
