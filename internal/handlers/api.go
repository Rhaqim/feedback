package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhaqim/worldgame/internal/game"
	"github.com/rhaqim/worldgame/internal/models"
)

type APIHandler struct {
	gameManager *game.GameManager
}

func NewAPIHandler(gm *game.GameManager) *APIHandler {
	return &APIHandler{gameManager: gm}
}

// POST /api/games - Create a game (no player created)
func (h *APIHandler) CreateGame(c *gin.Context) {
	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	g, err := h.gameManager.CreateGame(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[API] Game created: %s (%s)", g.ID, req.Name)

	c.JSON(http.StatusCreated, gin.H{
		"game_id": g.ID,
		"game":    g,
	})
}

// GET /api/games - List available games
func (h *APIHandler) ListGames(c *gin.Context) {
	games := h.gameManager.ListGames()
	c.JSON(http.StatusOK, gin.H{
		"games": games,
	})
}

// GET /api/games/:id - Get game details
func (h *APIHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")

	g, err := h.gameManager.GetGame(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	c.JSON(http.StatusOK, g)
}

// POST /api/games/:id/join - Join a game as a player
func (h *APIHandler) JoinGame(c *gin.Context) {
	gameID := c.Param("id")

	var req models.JoinGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.PlayerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_name is required"})
		return
	}
	if req.RegionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "region_id is required"})
		return
	}

	g, playerID, err := h.gameManager.JoinGame(gameID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[API] Player %s (%s) joined game %s", req.PlayerName, playerID, gameID)

	c.JSON(http.StatusOK, gin.H{
		"game_id":   g.ID,
		"player_id": playerID,
		"game":      g,
	})
}

// POST /api/games/:id/start - Start the game (host only)
func (h *APIHandler) StartGame(c *gin.Context) {
	gameID := c.Param("id")

	var body struct {
		PlayerID string `json:"player_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if body.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_id is required"})
		return
	}

	if err := h.gameManager.StartGame(gameID, body.PlayerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g, _ := h.gameManager.GetGame(gameID)
	log.Printf("[API] Game %s started", gameID)

	c.JSON(http.StatusOK, gin.H{
		"status": "started",
		"game":   g,
	})
}

// GET /api/regions - List all available regions
func (h *APIHandler) ListRegions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"regions": game.AllRegions,
	})
}
