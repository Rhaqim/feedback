package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
	"github.com/rhaqim/worldgame/internal/store"
)

// errorf is a convenience wrapper around fmt.Errorf used throughout the
// game package.
func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// GameManager is the thread-safe game orchestrator backed by PostgreSQL.
type GameManager struct {
	mu     sync.RWMutex
	store  *store.Store
	engine *Engine
	nextID int
}

// NewGameManager creates a new GameManager backed by the given engine and store.
func NewGameManager(engine *Engine, s *store.Store) *GameManager {
	return &GameManager{
		store:  s,
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
	ctx := context.Background()

	// Validate region from the database.
	region, err := gm.store.GetRegionByID(ctx, req.RegionID)
	if err != nil || region == nil {
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
	// Must be called before CreateGame to populate WeekStart/WeekEnd.
	gm.engine.InitializeGame(ctx, g)

	// Persist game to database.
	if err := gm.store.CreateGame(ctx, g); err != nil {
		return nil, errorf("failed to persist game: %v", err)
	}

	return g, nil
}

// JoinGame adds a player to an existing game. The first player to join
// becomes the host. No region selection -- everyone plays in the game's region.
func (gm *GameManager) JoinGame(gameID string, req models.JoinGameRequest) (*models.Game, string, error) {
	ctx := context.Background()

	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, err := gm.store.GetGame(ctx, gameID)
	if err != nil {
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

	// Persist player to database.
	if err := gm.store.CreatePlayer(ctx, gameID, player); err != nil {
		return nil, "", errorf("failed to persist player: %v", err)
	}

	game.Players[playerID] = player

	// First player to join becomes the host.
	if game.HostID == "" {
		game.HostID = playerID
		if err := gm.store.UpdateGameHost(ctx, gameID, playerID); err != nil {
			return nil, "", errorf("failed to set host: %v", err)
		}
	}

	return game, playerID, nil
}

// SubmitProposal delegates to the engine for proposal validation and recording.
func (gm *GameManager) SubmitProposal(gameID string, req models.SubmitProposalRequest) (*models.Proposal, error) {
	ctx := context.Background()

	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, err := gm.store.GetGame(ctx, gameID)
	if err != nil {
		return nil, errorf("game not found")
	}

	if game.Phase != models.PhaseActive {
		return nil, errorf("game is not in active phase")
	}

	return gm.engine.SubmitProposal(ctx, game, req)
}

// Evaluate triggers AI evaluation for the game (host only).
func (gm *GameManager) Evaluate(gameID, playerID string) (*models.Game, error) {
	ctx := context.Background()

	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, err := gm.store.GetGame(ctx, gameID)
	if err != nil {
		return nil, errorf("game not found")
	}

	if err := gm.engine.Evaluate(ctx, game, playerID); err != nil {
		return nil, err
	}

	return game, nil
}

// NextWeek starts the next week cycle (host only).
func (gm *GameManager) NextWeek(gameID, playerID string) (*models.Game, error) {
	ctx := context.Background()

	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, err := gm.store.GetGame(ctx, gameID)
	if err != nil {
		return nil, errorf("game not found")
	}

	if err := gm.engine.NextWeek(ctx, game, playerID); err != nil {
		return nil, err
	}

	return game, nil
}

// GetGame returns a game by ID.
func (gm *GameManager) GetGame(gameID string) (*models.Game, error) {
	ctx := context.Background()

	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, err := gm.store.GetGame(ctx, gameID)
	if err != nil {
		return nil, errorf("game not found")
	}
	return game, nil
}

// ListGames returns summaries of all games.
func (gm *GameManager) ListGames() []models.GameSummary {
	ctx := context.Background()

	gm.mu.RLock()
	defer gm.mu.RUnlock()

	summaries, err := gm.store.ListGames(ctx)
	if err != nil {
		return []models.GameSummary{}
	}
	return summaries
}

// SetPlayerConnected updates a player's connected status in the database.
func (gm *GameManager) SetPlayerConnected(gameID, playerID string, connected bool) {
	ctx := context.Background()

	gm.mu.Lock()
	defer gm.mu.Unlock()

	_ = gm.store.SetPlayerConnected(ctx, playerID, connected)
}
