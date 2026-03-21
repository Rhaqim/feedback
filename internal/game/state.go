package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
)

// GameManager is the thread-safe in-memory store for all active games.
type GameManager struct {
	mu     sync.RWMutex
	games  map[string]*models.Game
	engine *Engine
	nextID int
}

func NewGameManager(engine *Engine) *GameManager {
	return &GameManager{
		games:  make(map[string]*models.Game),
		engine: engine,
	}
}

func (gm *GameManager) generateID() string {
	gm.nextID++
	return fmt.Sprintf("game_%d_%d", time.Now().Unix(), gm.nextID)
}

// CreateGame creates a new game in lobby phase with NO players.
func (gm *GameManager) CreateGame(req models.CreateGameRequest) (*models.Game, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gameID := gm.generateID()

	maxTurns := req.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 30
	}

	maxPlayers := req.MaxPlayers
	if maxPlayers <= 0 {
		maxPlayers = 4
	}
	if maxPlayers < 2 {
		maxPlayers = 2
	}
	if maxPlayers > 6 {
		maxPlayers = 6
	}

	g := &models.Game{
		ID:           gameID,
		Name:         req.Name,
		MaxPlayers:   maxPlayers,
		Players:      make(map[string]*models.PlayerState),
		Phase:        models.PhaseLobby,
		CurrentTurn:  0,
		MaxTurns:     maxTurns,
		Events:       []models.WorldEvent{},
		ActiveEvents: []models.WorldEvent{},
		CreatedAt:    time.Now(),
		HostID:       "", // no host until a player joins
	}

	gm.games[gameID] = g
	return g, nil
}

// JoinGame adds a player to an existing game in lobby phase.
// The first player to join becomes the host.
func (gm *GameManager) JoinGame(gameID string, req models.JoinGameRequest) (*models.Game, string, error) {
	region := GetRegionByID(req.RegionID)
	if region == nil {
		return nil, "", fmt.Errorf("invalid region ID: %s", req.RegionID)
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return nil, "", fmt.Errorf("game not found")
	}

	if game.Phase != models.PhaseLobby {
		return nil, "", fmt.Errorf("game is not accepting new players")
	}

	if len(game.Players) >= game.MaxPlayers {
		return nil, "", fmt.Errorf("game is full (max %d players)", game.MaxPlayers)
	}

	// Check if region is already taken
	for _, ps := range game.Players {
		if ps.Player.RegionID == req.RegionID {
			return nil, "", fmt.Errorf("region %s is already taken", req.RegionID)
		}
	}

	playerID := fmt.Sprintf("player_%d_%d", time.Now().UnixNano(), len(game.Players))

	player := models.Player{
		ID:        playerID,
		Name:      req.PlayerName,
		RegionID:  req.RegionID,
		Resources: models.Resources{},
		Ready:     false,
		Connected: false,
	}

	playerState := &models.PlayerState{
		Player:  player,
		Sectors: make(map[models.SectorType]*models.SectorState),
		Score:   0,
	}

	game.Players[playerID] = playerState

	// First player to join becomes the host
	if game.HostID == "" {
		game.HostID = playerID
	}

	return game, playerID, nil
}

// StartGame transitions the game from lobby to playing.
func (gm *GameManager) StartGame(gameID, playerID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return fmt.Errorf("game not found")
	}

	if game.HostID != playerID {
		return fmt.Errorf("only the host can start the game")
	}

	return gm.engine.StartGame(game)
}

// GetGame returns a game by ID.
func (gm *GameManager) GetGame(gameID string) (*models.Game, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, ok := gm.games[gameID]
	if !ok {
		return nil, fmt.Errorf("game not found")
	}
	return game, nil
}

// ListGames returns summaries of all games.
func (gm *GameManager) ListGames() []models.GameSummary {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	summaries := make([]models.GameSummary, 0)
	for _, g := range gm.games {
		summaries = append(summaries, models.GameSummary{
			ID:          g.ID,
			Name:        g.Name,
			Phase:       g.Phase,
			PlayerCount: len(g.Players),
			MaxPlayers:  g.MaxPlayers,
			MaxTurns:    g.MaxTurns,
			CurrentTurn: g.CurrentTurn,
			CreatedAt:   g.CreatedAt,
		})
	}
	return summaries
}

// SubmitAction records a player action in the game.
func (gm *GameManager) SubmitAction(gameID string, action models.Action) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return fmt.Errorf("game not found")
	}

	return gm.engine.SubmitAction(game, action)
}

// SetPlayerReady marks a player as ready and processes the turn if all ready.
func (gm *GameManager) SetPlayerReady(gameID, playerID string) (bool, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return false, fmt.Errorf("game not found")
	}

	return gm.engine.SetPlayerReady(game, playerID)
}

// SetPlayerConnected updates the connected status of a player.
func (gm *GameManager) SetPlayerConnected(gameID, playerID string, connected bool) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return
	}

	ps, ok := game.Players[playerID]
	if !ok {
		return
	}

	ps.Player.Connected = connected
}
