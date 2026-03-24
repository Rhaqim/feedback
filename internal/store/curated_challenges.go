package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// CreateCuratedChallenge inserts a new curated challenge and returns its ID.
func (s *Store) CreateCuratedChallenge(ctx context.Context, cc models.CuratedChallenge) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx,
		`INSERT INTO curated_challenges (feed_item_id, tag, region_id, title, description, source, source_url, severity, active, curator_notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id`,
		cc.FeedItemID, string(cc.Tag), cc.RegionID, cc.Title, cc.Description,
		cc.Source, cc.SourceURL, cc.Severity, cc.Active, cc.CuratorNotes).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert curated challenge: %w", err)
	}
	return id, nil
}

// GetCuratedChallenges returns curated challenges with optional tag and region filters.
func (s *Store) GetCuratedChallenges(ctx context.Context, tag string, regionID string, activeOnly bool) ([]models.CuratedChallenge, error) {
	query := `SELECT id, feed_item_id, tag, region_id, title, description, source, source_url, severity, active, used_in_game, curator_notes, created_at
		 FROM curated_challenges WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if tag != "" {
		query += fmt.Sprintf(" AND tag = $%d", argIdx)
		args = append(args, tag)
		argIdx++
	}
	if regionID != "" {
		query += fmt.Sprintf(" AND (region_id = $%d OR region_id = '')", argIdx)
		args = append(args, regionID)
		argIdx++
	}
	if activeOnly {
		query += " AND active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query curated challenges: %w", err)
	}
	defer rows.Close()

	var challenges []models.CuratedChallenge
	for rows.Next() {
		var cc models.CuratedChallenge
		if err := rows.Scan(&cc.ID, &cc.FeedItemID, &cc.Tag, &cc.RegionID, &cc.Title, &cc.Description,
			&cc.Source, &cc.SourceURL, &cc.Severity, &cc.Active, &cc.UsedInGame, &cc.CuratorNotes, &cc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan curated challenge: %w", err)
		}
		challenges = append(challenges, cc)
	}
	return challenges, rows.Err()
}

// GetUnusedCuratedChallenges returns active curated challenges not yet used in a game,
// filtered by tag and optionally region. Returns up to limit items.
func (s *Store) GetUnusedCuratedChallenges(ctx context.Context, tag models.Tag, regionID string, limit int) ([]models.CuratedChallenge, error) {
	query := `SELECT id, feed_item_id, tag, region_id, title, description, source, source_url, severity, active, used_in_game, curator_notes, created_at
		 FROM curated_challenges
		 WHERE tag = $1 AND active = true AND used_in_game = false`
	args := []interface{}{string(tag)}

	if regionID != "" {
		query += ` AND (region_id = $2 OR region_id = '')`
		args = append(args, regionID)
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query unused curated challenges: %w", err)
	}
	defer rows.Close()

	var challenges []models.CuratedChallenge
	for rows.Next() {
		var cc models.CuratedChallenge
		if err := rows.Scan(&cc.ID, &cc.FeedItemID, &cc.Tag, &cc.RegionID, &cc.Title, &cc.Description,
			&cc.Source, &cc.SourceURL, &cc.Severity, &cc.Active, &cc.UsedInGame, &cc.CuratorNotes, &cc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan curated challenge: %w", err)
		}
		challenges = append(challenges, cc)
	}
	return challenges, rows.Err()
}

// MarkCuratedChallengeUsed sets used_in_game = true for a curated challenge.
func (s *Store) MarkCuratedChallengeUsed(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE curated_challenges SET used_in_game = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark curated challenge used: %w", err)
	}
	return nil
}

// DeleteCuratedChallenge removes a curated challenge by ID.
func (s *Store) DeleteCuratedChallenge(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM curated_challenges WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete curated challenge: %w", err)
	}
	return nil
}
