package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
)

// errorf is a convenience wrapper around fmt.Errorf used throughout the
// game package.
func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// GameManager is the thread-safe in-memory store for all active games.
type GameManager struct {
	mu     sync.RWMutex
	games  map[string]*models.Game
	engine *Engine
	nextID int
}

// NewGameManager creates a new GameManager backed by the given engine.
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

// CreateGame creates a new game, generates initial challenges, and puts it
// in the active phase. The creator does NOT automatically become a player.
func (gm *GameManager) CreateGame(req models.CreateGameRequest) (*models.Game, error) {
	// Validate region.
	region := GetRegionByID(req.RegionID)
	if region == nil {
		return nil, errorf("invalid region_id: %s", req.RegionID)
	}

	// Validate tags.
	if len(req.Tags) == 0 {
		return nil, errorf("at least one tag is required")
	}
	for _, t := range req.Tags {
		if !models.IsValidTag(t) {
			return nil, errorf("invalid tag: %s", t)
		}
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	gameID := gm.generateID()

	g := &models.Game{
		ID:         gameID,
		Name:       req.Name,
		RegionID:   region.ID,
		RegionName: region.Name,
		Tags:       req.Tags,
		Phase:      models.PhaseActive,
		Players:    make(map[string]*models.Player),
		Challenges: []models.Challenge{},
		Proposals:  []models.Proposal{},
		WeekNumber: 0, // InitializeGame will set to 1
		HostID:     "",
		CreatedAt:  time.Now(),
	}

	// Engine sets up week timing, generates challenges, etc.
	gm.engine.InitializeGame(g)

	gm.games[gameID] = g
	return g, nil
}

// JoinGame adds a player to an existing game. The first player to join
// becomes the host. No region selection -- everyone plays in the game's region.
func (gm *GameManager) JoinGame(gameID string, req models.JoinGameRequest) (*models.Game, string, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return nil, "", errorf("game not found")
	}

	playerID := fmt.Sprintf("player_%d_%d", time.Now().UnixNano(), len(game.Players))

	player := &models.Player{
		ID:         playerID,
		Name:       req.PlayerName,
		Points:     InitialPlayerPoints,
		TotalScore: 0,
		Connected:  false,
	}

	game.Players[playerID] = player

	// First player to join becomes the host.
	if game.HostID == "" {
		game.HostID = playerID
	}

	return game, playerID, nil
}

// SubmitProposal delegates to the engine for proposal validation and recording.
func (gm *GameManager) SubmitProposal(gameID string, req models.SubmitProposalRequest) (*models.Proposal, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return nil, errorf("game not found")
	}

	if game.Phase != models.PhaseActive {
		return nil, errorf("game is not in active phase")
	}

	return gm.engine.SubmitProposal(game, req)
}

// Evaluate triggers AI evaluation for the game (host only).
func (gm *GameManager) Evaluate(gameID, playerID string) (*models.Game, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return nil, errorf("game not found")
	}

	if err := gm.engine.Evaluate(game, playerID); err != nil {
		return nil, err
	}

	return game, nil
}

// NextWeek starts the next week cycle (host only).
func (gm *GameManager) NextWeek(gameID, playerID string) (*models.Game, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return nil, errorf("game not found")
	}

	if err := gm.engine.NextWeek(game, playerID); err != nil {
		return nil, err
	}

	return game, nil
}

// GetGame returns a game by ID.
func (gm *GameManager) GetGame(gameID string) (*models.Game, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, ok := gm.games[gameID]
	if !ok {
		return nil, errorf("game not found")
	}
	return game, nil
}

// ListGames returns summaries of all games.
func (gm *GameManager) ListGames() []models.GameSummary {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	summaries := make([]models.GameSummary, 0, len(gm.games))
	for _, g := range gm.games {
		summaries = append(summaries, models.GameSummary{
			ID:          g.ID,
			Name:        g.Name,
			RegionName:  g.RegionName,
			Tags:        g.Tags,
			Phase:       g.Phase,
			PlayerCount: len(g.Players),
			WeekNumber:  g.WeekNumber,
			CreatedAt:   g.CreatedAt,
		})
	}
	return summaries
}

// SetPlayerConnected updates a player's connected status.
func (gm *GameManager) SetPlayerConnected(gameID, playerID string, connected bool) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gameID]
	if !ok {
		return
	}

	player, ok := game.Players[playerID]
	if !ok {
		return
	}

	player.Connected = connected
}
