package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// CreateChatMessage inserts a new chat message.
func (s *Store) CreateChatMessage(ctx context.Context, gameID, playerID, playerName, message string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO chat_messages (game_id, player_id, player_name, message)
		 VALUES ($1, $2, $3, $4)`,
		gameID, playerID, playerName, message)
	if err != nil {
		return fmt.Errorf("insert chat message: %w", err)
	}
	return nil
}

// GetGameChat returns the most recent chat messages for a game, ordered oldest first.
func (s *Store) GetGameChat(ctx context.Context, gameID string, limit int) ([]models.ChatMessage, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, game_id, player_id, player_name, message, created_at
		 FROM chat_messages WHERE game_id = $1
		 ORDER BY created_at DESC LIMIT $2`, gameID, limit)
	if err != nil {
		return nil, fmt.Errorf("query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var m models.ChatMessage
		if err := rows.Scan(&m.ID, &m.GameID, &m.PlayerID, &m.PlayerName, &m.Message, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reverse to get oldest first (we queried DESC for LIMIT).
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	if messages == nil {
		messages = []models.ChatMessage{}
	}
	return messages, nil
}
