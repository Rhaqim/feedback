package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhaqim/worldgame/internal/game"
	"github.com/rhaqim/worldgame/internal/models"
	"github.com/rhaqim/worldgame/internal/store"
)

// APIHandler exposes REST endpoints for game management.
type APIHandler struct {
	gameManager *game.GameManager
	store       *store.Store
}

// NewAPIHandler creates a new APIHandler.
func NewAPIHandler(gm *game.GameManager, s *store.Store) *APIHandler {
	return &APIHandler{gameManager: gm, store: s}
}

// POST /api/games -- Create a game.
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
	if req.RegionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "region_id is required"})
		return
	}
	if len(req.Tags) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one tag is required"})
		return
	}

	g, err := h.gameManager.CreateGame(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[API] Game created: %s (%s) region=%s tags=%v", g.ID, req.Name, req.RegionID, req.Tags)

	c.JSON(http.StatusCreated, gin.H{
		"game_id": g.ID,
		"game":    g,
	})
}

// GET /api/games -- List all games.
func (h *APIHandler) ListGames(c *gin.Context) {
	games := h.gameManager.ListGames()
	c.JSON(http.StatusOK, gin.H{
		"games": games,
	})
}

// GET /api/games/:id -- Get a specific game.
func (h *APIHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")

	g, err := h.gameManager.GetGame(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	c.JSON(http.StatusOK, g)
}

// POST /api/games/:id/join -- Join a game.
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

// POST /api/games/:id/proposals -- Submit a proposal.
func (h *APIHandler) SubmitProposal(c *gin.Context) {
	gameID := c.Param("id")

	var req models.SubmitProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_id is required"})
		return
	}
	if req.ChallengeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "challenge_id is required"})
		return
	}
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}
	if req.PointsInvested <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "points_invested must be greater than 0"})
		return
	}

	proposal, err := h.gameManager.SubmitProposal(gameID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[API] Proposal submitted in game %s by player %s", gameID, req.PlayerID)

	c.JSON(http.StatusCreated, gin.H{
		"proposal": proposal,
	})
}

// POST /api/games/:id/evaluate -- Trigger AI evaluation (host only).
func (h *APIHandler) Evaluate(c *gin.Context) {
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

	g, err := h.gameManager.Evaluate(gameID, body.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[API] Game %s evaluated", gameID)

	c.JSON(http.StatusOK, gin.H{
		"status": "evaluated",
		"game":   g,
	})
}

// POST /api/games/:id/next-week -- Start next week (host only).
func (h *APIHandler) NextWeek(c *gin.Context) {
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

	g, err := h.gameManager.NextWeek(gameID, body.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[API] Game %s started week %d", gameID, g.WeekNumber)

	c.JSON(http.StatusOK, gin.H{
		"status": "started",
		"game":   g,
	})
}

// GET /api/regions -- List all available regions from the database.
func (h *APIHandler) ListRegions(c *gin.Context) {
	ctx := context.Background()
	regions, err := h.store.GetAllRegions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load regions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"regions": regions,
	})
}

// GET /api/games/:id/chat -- Get chat history for a game.
func (h *APIHandler) GetChat(c *gin.Context) {
	ctx := context.Background()
	gameID := c.Param("id")

	messages, err := h.store.GetGameChat(ctx, gameID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

// POST /api/regions -- Admin: add a new region.
func (h *APIHandler) CreateRegion(c *gin.Context) {
	ctx := context.Background()
	var r models.Region
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if r.ID == "" || r.Name == "" || r.Country == "" || r.Continent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id, name, country, and continent are required"})
		return
	}

	if err := h.store.UpsertRegion(ctx, r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create region"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"region": r})
}

// POST /api/challenge-templates -- Admin: add a challenge template.
func (h *APIHandler) CreateChallengeTemplate(c *gin.Context) {
	ctx := context.Background()
	var t models.ChallengeTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if t.Tag == "" || t.TitleTemplate == "" || t.DescriptionTemplate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag, title_template, and description_template are required"})
		return
	}
	if !models.IsValidTag(t.Tag) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag"})
		return
	}
	if t.Source == "" {
		t.Source = "news"
	}

	if err := h.store.CreateChallengeTemplate(ctx, t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create challenge template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"template": t})
}

// GET /api/challenge-templates -- List all challenge templates.
func (h *APIHandler) ListChallengeTemplates(c *gin.Context) {
	ctx := context.Background()
	templates, err := h.store.GetAllChallengeTemplates(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load challenge templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}
