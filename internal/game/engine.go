package game

import (
	"context"
	"log"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
	"github.com/rhaqim/worldgame/internal/store"
)

const (
	WeekDuration        = 7 * 24 * time.Hour
	InitialPlayerPoints = 100.0
	ChallengesPerTag    = 3 // generate 2-3 per tag; we use 3 for the prototype
)

// BroadcastFunc is called by the engine whenever it needs to push state to
// connected WebSocket clients.
type BroadcastFunc func(gameID string, msg models.WSMessage)

// Engine orchestrates the weekly game cycle: challenge generation, proposal
// submission, AI evaluation, and week transitions.
type Engine struct {
	challengeGen *ChallengeGenerator
	evaluator    *Evaluator
	broadcast    BroadcastFunc
	store        *store.Store
}

// NewEngine creates a new Engine with the given challenge generator,
// evaluator, broadcast function, and store.
func NewEngine(challengeGen *ChallengeGenerator, evaluator *Evaluator, broadcast BroadcastFunc, s *store.Store) *Engine {
	return &Engine{
		challengeGen: challengeGen,
		evaluator:    evaluator,
		broadcast:    broadcast,
		store:        s,
	}
}

// InitializeGame sets up a newly created game: generates initial challenges,
// sets the week window, and moves the game to active phase.
func (e *Engine) InitializeGame(ctx context.Context, g *models.Game) {
	now := time.Now()
	g.Phase = models.PhaseActive
	g.WeekNumber = 1
	g.WeekStart = now
	g.WeekEnd = now.Add(WeekDuration)
	g.Challenges = e.challengeGen.GenerateChallenges(ctx, g, ChallengesPerTag)
	g.Proposals = []models.Proposal{}
	g.Winner = nil

	// Persist challenges to the database.
	if err := e.store.CreateChallenges(ctx, g.ID, g.Challenges); err != nil {
		log.Printf("[Engine] Error persisting challenges for game %s: %v", g.ID, err)
	}

	log.Printf("[Engine] Game %s initialized with %d challenges", g.ID, len(g.Challenges))
}

// SubmitProposal validates and records a player's proposal for a challenge.
// It deducts the invested points from the player and broadcasts the proposal.
func (e *Engine) SubmitProposal(ctx context.Context, g *models.Game, req models.SubmitProposalRequest) (*models.Proposal, error) {
	player, ok := g.Players[req.PlayerID]
	if !ok {
		return nil, errorf("player %s not found in game", req.PlayerID)
	}

	// Find the challenge.
	var challenge *models.Challenge
	for i := range g.Challenges {
		if g.Challenges[i].ID == req.ChallengeID {
			challenge = &g.Challenges[i]
			break
		}
	}
	if challenge == nil {
		return nil, errorf("challenge %s not found", req.ChallengeID)
	}
	if !challenge.Active {
		return nil, errorf("challenge %s is no longer active", req.ChallengeID)
	}

	if req.PointsInvested <= 0 {
		return nil, errorf("points_invested must be greater than 0")
	}
	if player.Points < req.PointsInvested {
		return nil, errorf("insufficient points: have %.1f, need %.1f", player.Points, req.PointsInvested)
	}
	if req.Description == "" {
		return nil, errorf("description is required")
	}

	// Deduct points.
	player.Points -= req.PointsInvested

	proposal := models.Proposal{
		ID:             generateProposalID(g.ID, req.PlayerID),
		PlayerID:       req.PlayerID,
		PlayerName:     player.Name,
		ChallengeID:    req.ChallengeID,
		Description:    req.Description,
		PointsInvested: req.PointsInvested,
		SubmittedAt:    time.Now(),
		AIScore:        0,
		AIFeedback:     "",
	}

	g.Proposals = append(g.Proposals, proposal)

	// Persist to database.
	if err := e.store.CreateProposal(ctx, g.ID, proposal); err != nil {
		log.Printf("[Engine] Error persisting proposal: %v", err)
	}
	if err := e.store.UpdatePlayerPoints(ctx, req.PlayerID, player.Points); err != nil {
		log.Printf("[Engine] Error updating player points: %v", err)
	}

	// Broadcast the new proposal to all connected players.
	e.broadcast(g.ID, models.WSMessage{
		Type:    "proposal_submitted",
		Payload: proposal,
	})

	log.Printf("[Engine] Player %s submitted proposal for challenge %s in game %s (%.1f points invested)",
		req.PlayerID, req.ChallengeID, g.ID, req.PointsInvested)

	return &proposal, nil
}

// Evaluate runs the AI evaluator on all proposals, determines the winner,
// and transitions the game to the completed phase.
func (e *Engine) Evaluate(ctx context.Context, g *models.Game, requestingPlayerID string) error {
	if g.HostID != requestingPlayerID {
		return errorf("only the host can trigger evaluation")
	}
	if g.Phase != models.PhaseActive {
		return errorf("game must be in active phase to evaluate")
	}

	g.Phase = models.PhaseEvaluating

	winner := e.evaluator.EvaluateProposals(g)
	g.Winner = winner

	// Persist proposal scores to database.
	for _, p := range g.Proposals {
		if err := e.store.UpdateProposalScore(ctx, p.ID, p.AIScore, p.AIFeedback); err != nil {
			log.Printf("[Engine] Error updating proposal score: %v", err)
		}
	}

	// Add winning score to the winner's total.
	if winner != nil {
		if p, ok := g.Players[winner.PlayerID]; ok {
			p.TotalScore += winner.Score
			if err := e.store.UpdatePlayerTotalScore(ctx, winner.PlayerID, p.TotalScore); err != nil {
				log.Printf("[Engine] Error updating winner total score: %v", err)
			}
		}
	}

	g.Phase = models.PhaseCompleted

	// Persist game state.
	if err := e.store.UpdateGamePhase(ctx, g.ID, g.Phase); err != nil {
		log.Printf("[Engine] Error updating game phase: %v", err)
	}
	if err := e.store.UpdateGameWinner(ctx, g.ID, winner); err != nil {
		log.Printf("[Engine] Error updating game winner: %v", err)
	}

	// Broadcast evaluation results.
	e.broadcast(g.ID, models.WSMessage{
		Type: "evaluation_result",
		Payload: map[string]interface{}{
			"winner":    winner,
			"proposals": g.Proposals,
		},
	})

	// Broadcast updated game state.
	e.broadcastGameState(g)

	log.Printf("[Engine] Game %s week %d evaluated. Winner: %v", g.ID, g.WeekNumber, winner)
	return nil
}

// NextWeek resets the game for a new weekly cycle. The game must be in the
// completed phase. Player points are reset, proposals cleared, new challenges
// generated, and cumulative total_score is preserved.
func (e *Engine) NextWeek(ctx context.Context, g *models.Game, requestingPlayerID string) error {
	if g.HostID != requestingPlayerID {
		return errorf("only the host can start the next week")
	}
	if g.Phase != models.PhaseCompleted {
		return errorf("game must be in completed phase to start next week")
	}

	// Reset for new week.
	g.Phase = models.PhaseActive
	g.WeekNumber++
	now := time.Now()
	g.WeekStart = now
	g.WeekEnd = now.Add(WeekDuration)
	g.Winner = nil

	// Reset player points.
	for _, p := range g.Players {
		p.Points = InitialPlayerPoints
	}

	// Clear old proposals.
	g.Proposals = []models.Proposal{}

	// Generate new challenges.
	g.Challenges = e.challengeGen.GenerateChallenges(ctx, g, ChallengesPerTag)

	// Persist all changes to database.
	if err := e.store.ResetPlayerPoints(ctx, g.ID, InitialPlayerPoints); err != nil {
		log.Printf("[Engine] Error resetting player points: %v", err)
	}
	if err := e.store.ClearGameProposals(ctx, g.ID); err != nil {
		log.Printf("[Engine] Error clearing proposals: %v", err)
	}
	if err := e.store.DeactivateGameChallenges(ctx, g.ID); err != nil {
		log.Printf("[Engine] Error deactivating challenges: %v", err)
	}
	if err := e.store.CreateChallenges(ctx, g.ID, g.Challenges); err != nil {
		log.Printf("[Engine] Error creating new challenges: %v", err)
	}
	if err := e.store.UpdateGameWeek(ctx, g); err != nil {
		log.Printf("[Engine] Error updating game week: %v", err)
	}

	// Broadcast new state.
	e.broadcastGameState(g)

	// Broadcast new challenges individually.
	for _, ch := range g.Challenges {
		e.broadcast(g.ID, models.WSMessage{
			Type:    "new_challenge",
			Payload: ch,
		})
	}

	log.Printf("[Engine] Game %s started week %d with %d challenges", g.ID, g.WeekNumber, len(g.Challenges))
	return nil
}

func (e *Engine) broadcastGameState(g *models.Game) {
	e.broadcast(g.ID, models.WSMessage{
		Type:    "game_state",
		Payload: g,
	})
}

// generateProposalID creates a unique proposal ID.
func generateProposalID(gameID, playerID string) string {
	return "prop_" + gameID + "_" + playerID + "_" + time.Now().Format("150405.000")
}
