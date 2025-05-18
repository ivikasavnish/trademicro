package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/vikasavnish/trademicro/internal/models"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	connections map[*websocket.Conn]bool

	// Messages to be broadcast to all connected clients
	broadcast chan models.Message

	// Upgrader for HTTP connections to WebSocket
	upgrader websocket.Upgrader
}

// NewHub creates a new hub for managing WebSocket connections
func NewHub() *Hub {
	upgrader := websocket.Upgrader{
		// Allow all origins for WebSocket connections
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	return &Hub{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan models.Message),
		upgrader:    upgrader,
	}
}

// Run starts listening for messages to broadcast
func (h *Hub) Run() {
	for {
		// Wait for a message to broadcast
		msg := <-h.broadcast

		// Send to all connected clients
		for client := range h.connections {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				client.Close()
				delete(h.connections, client)
			}
		}
	}
}

// HandleWebSocket upgrades an HTTP connection to WebSocket
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Register new client
	h.connections[ws] = true

	// Read messages from the client (to keep the connection alive)
	go func() {
		defer ws.Close()
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				delete(h.connections, ws)
				break
			}
		}
	}()
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(msg models.Message) {
	h.broadcast <- msg
}
