package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// CreateGame inserts a new game and its tags in a transaction.
func (s *Store) CreateGame(ctx context.Context, g *models.Game) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO games (id, name, region_id, region_name, phase, week_number, week_start, week_end, host_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		g.ID, g.Name, g.RegionID, g.RegionName, string(g.Phase),
		g.WeekNumber, g.WeekStart, g.WeekEnd, g.HostID, g.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert game: %w", err)
	}

	for _, tag := range g.Tags {
		_, err = tx.Exec(ctx,
			`INSERT INTO game_tags (game_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			g.ID, string(tag))
		if err != nil {
			return fmt.Errorf("insert game tag: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetGame loads a full game from the database including tags, players,
// challenges, and proposals.
func (s *Store) GetGame(ctx context.Context, id string) (*models.Game, error) {
	g := &models.Game{
		Players:    make(map[string]*models.Player),
		Challenges: []models.Challenge{},
		Proposals:  []models.Proposal{},
		Tags:       []models.Tag{},
	}

	// Load game row.
	var winnerPlayerID, winnerPlayerName, winnerSummary *string
	var winnerScore *float64
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, region_id, region_name, phase, week_number, week_start, week_end,
		        host_id, winner_player_id, winner_player_name, winner_score, winner_summary, created_at
		 FROM games WHERE id = $1`, id).
		Scan(&g.ID, &g.Name, &g.RegionID, &g.RegionName, &g.Phase,
			&g.WeekNumber, &g.WeekStart, &g.WeekEnd, &g.HostID,
			&winnerPlayerID, &winnerPlayerName, &winnerScore, &winnerSummary,
			&g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get game %s: %w", id, err)
	}

	if winnerPlayerID != nil && winnerPlayerName != nil && winnerScore != nil {
		g.Winner = &models.WeekWinner{
			PlayerID:   *winnerPlayerID,
			PlayerName: *winnerPlayerName,
			Score:      *winnerScore,
		}
		if winnerSummary != nil {
			g.Winner.Summary = *winnerSummary
		}
	}

	// Load tags.
	tagRows, err := s.pool.Query(ctx,
		`SELECT tag FROM game_tags WHERE game_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("query game tags: %w", err)
	}
	defer tagRows.Close()
	for tagRows.Next() {
		var tag models.Tag
		if err := tagRows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan game tag: %w", err)
		}
		g.Tags = append(g.Tags, tag)
	}
	if err := tagRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate game tags: %w", err)
	}

	// Load players.
	playerRows, err := s.pool.Query(ctx,
		`SELECT id, name, points, total_score, connected FROM players WHERE game_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("query players: %w", err)
	}
	defer playerRows.Close()
	for playerRows.Next() {
		p := &models.Player{}
		if err := playerRows.Scan(&p.ID, &p.Name, &p.Points, &p.TotalScore, &p.Connected); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		g.Players[p.ID] = p
	}
	if err := playerRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate players: %w", err)
	}

	// Load challenges.
	challengeRows, err := s.pool.Query(ctx,
		`SELECT id, tag, title, description, source, source_url, region, severity, active, created_at
		 FROM challenges WHERE game_id = $1 ORDER BY created_at`, id)
	if err != nil {
		return nil, fmt.Errorf("query challenges: %w", err)
	}
	defer challengeRows.Close()
	for challengeRows.Next() {
		var ch models.Challenge
		if err := challengeRows.Scan(&ch.ID, &ch.Tag, &ch.Title, &ch.Description,
			&ch.Source, &ch.SourceURL, &ch.Region, &ch.Severity, &ch.Active, &ch.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan challenge: %w", err)
		}
		g.Challenges = append(g.Challenges, ch)
	}
	if err := challengeRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate challenges: %w", err)
	}

	// Load proposals.
	proposalRows, err := s.pool.Query(ctx,
		`SELECT id, player_id, player_name, challenge_id, description,
		        points_invested, ai_score, ai_feedback, submitted_at
		 FROM proposals WHERE game_id = $1 ORDER BY submitted_at`, id)
	if err != nil {
		return nil, fmt.Errorf("query proposals: %w", err)
	}
	defer proposalRows.Close()
	for proposalRows.Next() {
		var p models.Proposal
		if err := proposalRows.Scan(&p.ID, &p.PlayerID, &p.PlayerName, &p.ChallengeID,
			&p.Description, &p.PointsInvested, &p.AIScore, &p.AIFeedback, &p.SubmittedAt); err != nil {
			return nil, fmt.Errorf("scan proposal: %w", err)
		}
		g.Proposals = append(g.Proposals, p)
	}
	if err := proposalRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate proposals: %w", err)
	}

	return g, nil
}

// ListGames returns lightweight summaries of all games.
func (s *Store) ListGames(ctx context.Context) ([]models.GameSummary, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT g.id, g.name, g.region_name, g.phase, g.week_number, g.created_at,
		        (SELECT COUNT(*) FROM players p WHERE p.game_id = g.id) AS player_count
		 FROM games g ORDER BY g.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query games: %w", err)
	}
	defer rows.Close()

	var summaries []models.GameSummary
	for rows.Next() {
		var gs models.GameSummary
		if err := rows.Scan(&gs.ID, &gs.Name, &gs.RegionName, &gs.Phase,
			&gs.WeekNumber, &gs.CreatedAt, &gs.PlayerCount); err != nil {
			return nil, fmt.Errorf("scan game summary: %w", err)
		}
		gs.Tags = []models.Tag{} // initialize
		summaries = append(summaries, gs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate games: %w", err)
	}

	// Load tags for each game.
	for i := range summaries {
		tagRows, err := s.pool.Query(ctx,
			`SELECT tag FROM game_tags WHERE game_id = $1`, summaries[i].ID)
		if err != nil {
			return nil, fmt.Errorf("query game tags: %w", err)
		}
		for tagRows.Next() {
			var tag models.Tag
			if err := tagRows.Scan(&tag); err != nil {
				tagRows.Close()
				return nil, fmt.Errorf("scan game tag: %w", err)
			}
			summaries[i].Tags = append(summaries[i].Tags, tag)
		}
		tagRows.Close()
		if err := tagRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate game tags: %w", err)
		}
	}

	if summaries == nil {
		summaries = []models.GameSummary{}
	}

	return summaries, nil
}

// UpdateGamePhase updates the phase of a game.
func (s *Store) UpdateGamePhase(ctx context.Context, gameID string, phase models.GamePhase) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE games SET phase = $1 WHERE id = $2`,
		string(phase), gameID)
	if err != nil {
		return fmt.Errorf("update game phase: %w", err)
	}
	return nil
}

// UpdateGameWeek updates week-related fields on a game (for NextWeek).
func (s *Store) UpdateGameWeek(ctx context.Context, g *models.Game) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE games SET week_number = $1, week_start = $2, week_end = $3, phase = $4,
		        winner_player_id = NULL, winner_player_name = NULL, winner_score = NULL, winner_summary = NULL
		 WHERE id = $5`,
		g.WeekNumber, g.WeekStart, g.WeekEnd, string(g.Phase), g.ID)
	if err != nil {
		return fmt.Errorf("update game week: %w", err)
	}
	return nil
}

// UpdateGameWinner sets the winner fields on a game.
func (s *Store) UpdateGameWinner(ctx context.Context, gameID string, winner *models.WeekWinner) error {
	if winner == nil {
		_, err := s.pool.Exec(ctx,
			`UPDATE games SET winner_player_id = NULL, winner_player_name = NULL,
			        winner_score = NULL, winner_summary = NULL WHERE id = $1`, gameID)
		return err
	}
	_, err := s.pool.Exec(ctx,
		`UPDATE games SET winner_player_id = $1, winner_player_name = $2,
		        winner_score = $3, winner_summary = $4 WHERE id = $5`,
		winner.PlayerID, winner.PlayerName, winner.Score, winner.Summary, gameID)
	if err != nil {
		return fmt.Errorf("update game winner: %w", err)
	}
	return nil
}

// UpdateGameHost sets the host_id on a game.
func (s *Store) UpdateGameHost(ctx context.Context, gameID, hostID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE games SET host_id = $1 WHERE id = $2`, hostID, gameID)
	if err != nil {
		return fmt.Errorf("update game host: %w", err)
	}
	return nil
}

// GameExists checks whether a game with the given ID exists.
func (s *Store) GameExists(ctx context.Context, gameID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM games WHERE id = $1)`, gameID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check game exists: %w", err)
	}
	return exists, nil
}
