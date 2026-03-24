package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// GetChallengeTemplates returns all challenge templates for a given tag.
func (s *Store) GetChallengeTemplates(ctx context.Context, tag models.Tag) ([]models.ChallengeTemplate, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, tag, title_template, description_template, source, created_at
		 FROM challenge_templates WHERE tag = $1`, string(tag))
	if err != nil {
		return nil, fmt.Errorf("query challenge templates: %w", err)
	}
	defer rows.Close()

	var templates []models.ChallengeTemplate
	for rows.Next() {
		var t models.ChallengeTemplate
		if err := rows.Scan(&t.ID, &t.Tag, &t.TitleTemplate, &t.DescriptionTemplate, &t.Source, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan challenge template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// GetAllChallengeTemplates returns all challenge templates across all tags.
func (s *Store) GetAllChallengeTemplates(ctx context.Context) ([]models.ChallengeTemplate, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, tag, title_template, description_template, source, created_at
		 FROM challenge_templates ORDER BY tag, id`)
	if err != nil {
		return nil, fmt.Errorf("query all challenge templates: %w", err)
	}
	defer rows.Close()

	var templates []models.ChallengeTemplate
	for rows.Next() {
		var t models.ChallengeTemplate
		if err := rows.Scan(&t.ID, &t.Tag, &t.TitleTemplate, &t.DescriptionTemplate, &t.Source, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan challenge template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// CreateChallengeTemplate inserts a new challenge template.
func (s *Store) CreateChallengeTemplate(ctx context.Context, t models.ChallengeTemplate) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO challenge_templates (tag, title_template, description_template, source)
		 VALUES ($1, $2, $3, $4)`,
		string(t.Tag), t.TitleTemplate, t.DescriptionTemplate, t.Source)
	if err != nil {
		return fmt.Errorf("insert challenge template: %w", err)
	}
	return nil
}

// CreateChallenge inserts a single challenge for a game.
func (s *Store) CreateChallenge(ctx context.Context, gameID string, c models.Challenge) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO challenges (id, game_id, tag, title, description, source, source_url, region, severity, active, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		c.ID, gameID, string(c.Tag), c.Title, c.Description, c.Source,
		c.SourceURL, c.Region, c.Severity, c.Active, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert challenge: %w", err)
	}
	return nil
}

// CreateChallenges inserts multiple challenges for a game in a transaction.
func (s *Store) CreateChallenges(ctx context.Context, gameID string, challenges []models.Challenge) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, c := range challenges {
		_, err := tx.Exec(ctx,
			`INSERT INTO challenges (id, game_id, tag, title, description, source, source_url, region, severity, active, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			c.ID, gameID, string(c.Tag), c.Title, c.Description, c.Source,
			c.SourceURL, c.Region, c.Severity, c.Active, c.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert challenge %s: %w", c.ID, err)
		}
	}

	return tx.Commit(ctx)
}

// GetGameChallenges returns all challenges for a game.
func (s *Store) GetGameChallenges(ctx context.Context, gameID string) ([]models.Challenge, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, tag, title, description, source, source_url, region, severity, active, created_at
		 FROM challenges WHERE game_id = $1 ORDER BY created_at`, gameID)
	if err != nil {
		return nil, fmt.Errorf("query challenges: %w", err)
	}
	defer rows.Close()

	var challenges []models.Challenge
	for rows.Next() {
		var ch models.Challenge
		if err := rows.Scan(&ch.ID, &ch.Tag, &ch.Title, &ch.Description,
			&ch.Source, &ch.SourceURL, &ch.Region, &ch.Severity, &ch.Active, &ch.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan challenge: %w", err)
		}
		challenges = append(challenges, ch)
	}
	if challenges == nil {
		challenges = []models.Challenge{}
	}
	return challenges, rows.Err()
}

// GetChallenge returns a single challenge by ID.
func (s *Store) GetChallenge(ctx context.Context, id string) (*models.Challenge, error) {
	var ch models.Challenge
	err := s.pool.QueryRow(ctx,
		`SELECT id, tag, title, description, source, source_url, region, severity, active, created_at
		 FROM challenges WHERE id = $1`, id).
		Scan(&ch.ID, &ch.Tag, &ch.Title, &ch.Description,
			&ch.Source, &ch.SourceURL, &ch.Region, &ch.Severity, &ch.Active, &ch.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get challenge %s: %w", id, err)
	}
	return &ch, nil
}

// DeactivateGameChallenges sets active=false for all challenges in a game.
func (s *Store) DeactivateGameChallenges(ctx context.Context, gameID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE challenges SET active = false WHERE game_id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("deactivate challenges: %w", err)
	}
	return nil
}
