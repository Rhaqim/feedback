package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rhaqim/worldgame/internal/game"
	"github.com/rhaqim/worldgame/internal/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins
	},
}

// Client represents a single WebSocket connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	gameID   string
	playerID string
}

// Hub manages all active WebSocket connections.
type Hub struct {
	mu          sync.RWMutex
	clients     map[*Client]bool
	gameClients map[string]map[*Client]bool // gameID -> clients
	register    chan *Client
	unregister  chan *Client
	gameManager *game.GameManager
}

func NewHub(gm *game.GameManager) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		gameClients: make(map[string]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		gameManager: gm,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if client.gameID != "" {
				if h.gameClients[client.gameID] == nil {
					h.gameClients[client.gameID] = make(map[*Client]bool)
				}
				h.gameClients[client.gameID][client] = true
			}
			h.mu.Unlock()
			log.Printf("[Hub] Client registered: player=%s game=%s", client.playerID, client.gameID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.gameID != "" {
					if gc, ok := h.gameClients[client.gameID]; ok {
						delete(gc, client)
					}
					h.gameManager.SetPlayerConnected(client.gameID, client.playerID, false)
				}
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[Hub] Client unregistered: player=%s game=%s", client.playerID, client.gameID)

			// Broadcast player left
			if client.gameID != "" {
				h.BroadcastToGame(client.gameID, models.WSMessage{
					Type: "player_left",
					Payload: map[string]string{
						"player_id": client.playerID,
					},
				})
			}
		}
	}
}

// BroadcastToGame sends a message to all clients in a specific game.
func (h *Hub) BroadcastToGame(gameID string, msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Hub] Error marshaling broadcast: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.gameClients[gameID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.send <- data:
		default:
			h.mu.Lock()
			delete(h.clients, client)
			if gc, ok := h.gameClients[gameID]; ok {
				delete(gc, client)
			}
			close(client.send)
			h.mu.Unlock()
		}
	}
}

// HandleWebSocket upgrades an HTTP connection to WebSocket.
// This uses the raw http.ResponseWriter and *http.Request so it works with
// gorilla/websocket. The gin handler in main.go will extract them.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Hub] Upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Client] Read error: %v", err)
			}
			break
		}
		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(raw []byte) {
	var msg models.WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("[Client] Invalid message: %v", err)
		c.sendError("invalid message format")
		return
	}

	switch msg.Type {
	case "join_game":
		c.handleJoinGame(msg.Payload)
	case "player_action":
		c.handlePlayerAction(msg.Payload)
	case "ready":
		c.handleReady()
	case "chat":
		c.handleChat(msg.Payload)
	default:
		c.sendError("unknown message type: " + msg.Type)
	}
}

func (c *Client) handleJoinGame(payload interface{}) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		c.sendError("invalid join_game payload")
		return
	}

	gameID, _ := data["game_id"].(string)
	playerID, _ := data["player_id"].(string)

	if gameID == "" || playerID == "" {
		c.sendError("game_id and player_id are required")
		return
	}

	// Verify game and player exist
	g, err := c.hub.gameManager.GetGame(gameID)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if _, exists := g.Players[playerID]; !exists {
		c.sendError("player not found in game")
		return
	}

	c.gameID = gameID
	c.playerID = playerID
	c.hub.register <- c
	c.hub.gameManager.SetPlayerConnected(gameID, playerID, true)

	// Send current game state
	stateMsg, _ := json.Marshal(models.WSMessage{
		Type:    "game_state",
		Payload: g,
	})
	c.send <- stateMsg

	// Broadcast player joined
	c.hub.BroadcastToGame(gameID, models.WSMessage{
		Type: "player_joined",
		Payload: map[string]string{
			"player_id": playerID,
		},
	})
}

func (c *Client) handlePlayerAction(payload interface{}) {
	if c.gameID == "" || c.playerID == "" {
		c.sendError("not joined to a game")
		return
	}

	data, ok := payload.(map[string]interface{})
	if !ok {
		c.sendError("invalid action payload")
		return
	}

	actionTypeStr, _ := data["type"].(string)
	actionData, _ := data["data"].(map[string]interface{})

	action := models.Action{
		Type:     models.ActionType(actionTypeStr),
		PlayerID: c.playerID,
		Data:     actionData,
	}

	if err := c.hub.gameManager.SubmitAction(c.gameID, action); err != nil {
		c.sendError(err.Error())
		return
	}

	// Acknowledge
	ack, _ := json.Marshal(models.WSMessage{
		Type:    "action_ack",
		Payload: map[string]string{"status": "ok"},
	})
	c.send <- ack
}

func (c *Client) handleReady() {
	if c.gameID == "" || c.playerID == "" {
		c.sendError("not joined to a game")
		return
	}

	turnProcessed, err := c.hub.gameManager.SetPlayerReady(c.gameID, c.playerID)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if !turnProcessed {
		// Just acknowledge readiness
		ack, _ := json.Marshal(models.WSMessage{
			Type:    "ready_ack",
			Payload: map[string]string{"status": "waiting"},
		})
		c.send <- ack
	}
	// If turn was processed, the engine will broadcast updates via the hub
}

func (c *Client) handleChat(payload interface{}) {
	if c.gameID == "" {
		c.sendError("not joined to a game")
		return
	}

	data, ok := payload.(map[string]interface{})
	if !ok {
		c.sendError("invalid chat payload")
		return
	}

	message, _ := data["message"].(string)
	if message == "" {
		return
	}

	// Look up player name from game state
	playerName := ""
	g, err := c.hub.gameManager.GetGame(c.gameID)
	if err == nil {
		if ps, exists := g.Players[c.playerID]; exists {
			playerName = ps.Player.Name
		}
	}

	c.hub.BroadcastToGame(c.gameID, models.WSMessage{
		Type: "chat",
		Payload: map[string]interface{}{
			"player_id":   c.playerID,
			"player_name": playerName,
			"message":     message,
			"timestamp":   time.Now().Unix(),
		},
	})
}

func (c *Client) sendError(msg string) {
	errMsg, _ := json.Marshal(models.WSMessage{
		Type:    "error",
		Payload: map[string]string{"message": msg},
	})
	c.send <- errMsg
}
