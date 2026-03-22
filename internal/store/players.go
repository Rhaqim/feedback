package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// CreatePlayer inserts a new player for a game.
func (s *Store) CreatePlayer(ctx context.Context, gameID string, p *models.Player) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO players (id, game_id, name, points, total_score, connected)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		p.ID, gameID, p.Name, p.Points, p.TotalScore, p.Connected)
	if err != nil {
		return fmt.Errorf("insert player: %w", err)
	}
	return nil
}

// GetPlayer returns a single player by ID.
func (s *Store) GetPlayer(ctx context.Context, id string) (*models.Player, error) {
	p := &models.Player{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, points, total_score, connected FROM players WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Points, &p.TotalScore, &p.Connected)
	if err != nil {
		return nil, fmt.Errorf("get player %s: %w", id, err)
	}
	return p, nil
}

// GetGamePlayers returns all players in a game as a map.
func (s *Store) GetGamePlayers(ctx context.Context, gameID string) (map[string]*models.Player, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, points, total_score, connected FROM players WHERE game_id = $1`, gameID)
	if err != nil {
		return nil, fmt.Errorf("query players: %w", err)
	}
	defer rows.Close()

	players := make(map[string]*models.Player)
	for rows.Next() {
		p := &models.Player{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Points, &p.TotalScore, &p.Connected); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		players[p.ID] = p
	}
	return players, rows.Err()
}

// UpdatePlayerPoints updates the available points for a player.
func (s *Store) UpdatePlayerPoints(ctx context.Context, playerID string, points float64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE players SET points = $1 WHERE id = $2`, points, playerID)
	if err != nil {
		return fmt.Errorf("update player points: %w", err)
	}
	return nil
}

// UpdatePlayerTotalScore updates the cumulative total score for a player.
func (s *Store) UpdatePlayerTotalScore(ctx context.Context, playerID string, totalScore float64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE players SET total_score = $1 WHERE id = $2`, totalScore, playerID)
	if err != nil {
		return fmt.Errorf("update player total score: %w", err)
	}
	return nil
}

// ResetPlayerPoints resets all players in a game to the given points value.
func (s *Store) ResetPlayerPoints(ctx context.Context, gameID string, points float64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE players SET points = $1 WHERE game_id = $2`, points, gameID)
	if err != nil {
		return fmt.Errorf("reset player points: %w", err)
	}
	return nil
}

// SetPlayerConnected updates a player's connected status.
func (s *Store) SetPlayerConnected(ctx context.Context, playerID string, connected bool) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE players SET connected = $1 WHERE id = $2`, connected, playerID)
	if err != nil {
		return fmt.Errorf("set player connected: %w", err)
	}
	return nil
}
