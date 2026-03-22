package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// CreateProposal inserts a new proposal for a game.
func (s *Store) CreateProposal(ctx context.Context, gameID string, p models.Proposal) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO proposals (id, game_id, player_id, player_name, challenge_id, description,
		        points_invested, ai_score, ai_feedback, submitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		p.ID, gameID, p.PlayerID, p.PlayerName, p.ChallengeID, p.Description,
		p.PointsInvested, p.AIScore, p.AIFeedback, p.SubmittedAt)
	if err != nil {
		return fmt.Errorf("insert proposal: %w", err)
	}
	return nil
}

// GetGameProposals returns all proposals for a game.
func (s *Store) GetGameProposals(ctx context.Context, gameID string) ([]models.Proposal, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, player_id, player_name, challenge_id, description,
		        points_invested, ai_score, ai_feedback, submitted_at
		 FROM proposals WHERE game_id = $1 ORDER BY submitted_at`, gameID)
	if err != nil {
		return nil, fmt.Errorf("query proposals: %w", err)
	}
	defer rows.Close()

	var proposals []models.Proposal
	for rows.Next() {
		var p models.Proposal
		if err := rows.Scan(&p.ID, &p.PlayerID, &p.PlayerName, &p.ChallengeID,
			&p.Description, &p.PointsInvested, &p.AIScore, &p.AIFeedback, &p.SubmittedAt); err != nil {
			return nil, fmt.Errorf("scan proposal: %w", err)
		}
		proposals = append(proposals, p)
	}
	if proposals == nil {
		proposals = []models.Proposal{}
	}
	return proposals, rows.Err()
}

// UpdateProposalScore updates the AI score and feedback for a proposal.
func (s *Store) UpdateProposalScore(ctx context.Context, proposalID string, aiScore float64, aiFeedback string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE proposals SET ai_score = $1, ai_feedback = $2 WHERE id = $3`,
		aiScore, aiFeedback, proposalID)
	if err != nil {
		return fmt.Errorf("update proposal score: %w", err)
	}
	return nil
}

// ClearGameProposals deletes all proposals for a game.
func (s *Store) ClearGameProposals(ctx context.Context, gameID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM proposals WHERE game_id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("clear proposals: %w", err)
	}
	return nil
}
