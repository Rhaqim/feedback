package game

import (
	"fmt"
	"log"
	"math"

	"github.com/rhaqim/worldgame/internal/models"
)

const (
	WinScore       = 400.0
	MaxScore       = 500.0
	DecayRate      = 1.5 // sectors lose this much per turn if not invested
	InvestmentRate = 0.5 // multiplier: budget spent → sector level gain
	MaxSectorLevel = 100.0
	MinSectorLevel = 0.0
)

// BroadcastFunc is called by the engine whenever it needs to push state to clients.
type BroadcastFunc func(gameID string, msg models.WSMessage)

type Engine struct {
	eventGen  *EventGenerator
	broadcast BroadcastFunc
}

func NewEngine(eventGen *EventGenerator, broadcast BroadcastFunc) *Engine {
	return &Engine{
		eventGen:  eventGen,
		broadcast: broadcast,
	}
}

// StartGame transitions a game from lobby to playing phase.
func (e *Engine) StartGame(g *models.Game) error {
	if g.Phase != models.PhaseLobby {
		return fmt.Errorf("game is not in lobby phase")
	}
	if len(g.Players) < 2 {
		return fmt.Errorf("need at least 2 players to start")
	}

	g.Phase = models.PhasePlaying
	g.CurrentTurn = 1

	// Initialize player sectors from region base stats
	for _, ps := range g.Players {
		region := GetRegionByID(ps.Player.RegionID)
		if region == nil {
			continue
		}
		ps.Sectors = make(map[models.SectorType]*models.SectorState)
		for _, s := range models.AllSectors {
			base := region.BaseStats[s]
			ps.Sectors[s] = &models.SectorState{
				Level:      base,
				Investment: 0,
				Growth:     1.0,
			}
		}
		ps.Player.Resources = models.Resources{
			Budget:    100,
			Influence: 50,
			Stability: 70,
			Knowledge: 40,
		}
		ps.Score = e.calcScore(ps)
	}

	// Generate initial events
	initialEvents := e.eventGen.GenerateEvents(g, 2)
	g.ActiveEvents = initialEvents
	g.Events = append(g.Events, initialEvents...)

	log.Printf("[Engine] Game %s started with %d players", g.ID, len(g.Players))

	e.broadcastGameState(g)
	e.broadcastEvents(g, initialEvents)

	return nil
}

// SubmitAction records a player action for the current turn.
func (e *Engine) SubmitAction(g *models.Game, action models.Action) error {
	if g.Phase != models.PhasePlaying {
		return fmt.Errorf("game is not in playing phase")
	}

	ps, ok := g.Players[action.PlayerID]
	if !ok {
		return fmt.Errorf("player %s not in game", action.PlayerID)
	}

	ps.TurnActions = append(ps.TurnActions, action)
	log.Printf("[Engine] Player %s submitted action %s in game %s", action.PlayerID, action.Type, g.ID)
	return nil
}

// SetPlayerReady marks a player as ready. If all players are ready, process the turn.
func (e *Engine) SetPlayerReady(g *models.Game, playerID string) (bool, error) {
	if g.Phase != models.PhasePlaying {
		return false, fmt.Errorf("game is not in playing phase")
	}

	ps, ok := g.Players[playerID]
	if !ok {
		return false, fmt.Errorf("player %s not in game", playerID)
	}

	ps.Player.Ready = true
	log.Printf("[Engine] Player %s is ready in game %s", playerID, g.ID)

	// Check if all connected players are ready
	allReady := true
	for _, p := range g.Players {
		if p.Player.Connected && !p.Player.Ready {
			allReady = false
			break
		}
	}

	if allReady {
		e.ProcessTurn(g)
		return true, nil
	}
	return false, nil
}

// ProcessTurn handles end-of-turn logic.
func (e *Engine) ProcessTurn(g *models.Game) {
	log.Printf("[Engine] Processing turn %d for game %s", g.CurrentTurn, g.ID)

	// 1. Apply player allocations
	for _, ps := range g.Players {
		e.applyAllocations(ps)
	}

	// 2. Apply active events
	e.applyActiveEvents(g)

	// 3. Apply decay to uninvested sectors
	for _, ps := range g.Players {
		e.applyDecay(ps)
	}

	// 4. Clamp sector levels
	for _, ps := range g.Players {
		for _, ss := range ps.Sectors {
			ss.Level = math.Max(MinSectorLevel, math.Min(MaxSectorLevel, ss.Level))
		}
	}

	// 5. Update scores
	for _, ps := range g.Players {
		ps.Score = e.calcScore(ps)
	}

	// 6. Replenish resources
	for _, ps := range g.Players {
		e.replenishResources(ps)
	}

	// 7. Tick event durations and remove expired
	activeEvents := make([]models.WorldEvent, 0)
	for i := range g.ActiveEvents {
		g.ActiveEvents[i].Duration--
		if g.ActiveEvents[i].Duration > 0 {
			activeEvents = append(activeEvents, g.ActiveEvents[i])
		}
	}
	g.ActiveEvents = activeEvents

	// 8. Generate new events (1-2 per turn)
	newEventCount := 1
	if e.eventGen.rng.Float64() < 0.5 {
		newEventCount = 2
	}
	newEvents := e.eventGen.GenerateEvents(g, newEventCount)
	g.ActiveEvents = append(g.ActiveEvents, newEvents...)
	g.Events = append(g.Events, newEvents...)

	// 9. Check win condition
	var winner *models.PlayerState
	for _, ps := range g.Players {
		if ps.Score >= WinScore {
			winner = ps
			break
		}
	}

	// 10. Advance turn
	g.CurrentTurn++

	// 11. Reset player readiness and actions
	for _, ps := range g.Players {
		ps.Player.Ready = false
		ps.TurnActions = nil
	}

	// 12. Check end conditions
	if winner != nil || (g.MaxTurns > 0 && g.CurrentTurn > g.MaxTurns) {
		g.Phase = models.PhaseFinished
		if winner == nil {
			// Find highest score
			var best *models.PlayerState
			for _, ps := range g.Players {
				if best == nil || ps.Score > best.Score {
					best = ps
				}
			}
			winner = best
		}
		log.Printf("[Engine] Game %s finished! Winner: %s (score: %.1f)", g.ID, winner.Player.Name, winner.Score)
		e.broadcast(g.ID, models.WSMessage{
			Type: "game_over",
			Payload: map[string]interface{}{
				"winner_id":     winner.Player.ID,
				"winner_name":   winner.Player.Name,
				"winner_region": winner.Player.RegionID,
				"score":         winner.Score,
			},
		})
	}

	// Broadcast updates
	e.broadcastTurnResult(g)
	e.broadcastGameState(g)
	e.broadcastEvents(g, newEvents)
}

func (e *Engine) applyAllocations(ps *models.PlayerState) {
	for _, action := range ps.TurnActions {
		switch action.Type {
		case models.ActionAllocate:
			e.handleAllocate(ps, action)
		case models.ActionPolicy:
			e.handlePolicy(ps, action)
		case models.ActionTrade:
			// Trade handled separately between two players
		case models.ActionRespondEvent:
			e.handleEventResponse(ps, action)
		}
	}
}

func (e *Engine) handleAllocate(ps *models.PlayerState, action models.Action) {
	// Data: {"sector": "economics", "amount": 20}
	sectorStr, ok := action.Data["sector"].(string)
	if !ok {
		return
	}
	amount, ok := action.Data["amount"].(float64)
	if !ok {
		return
	}

	sector := models.SectorType(sectorStr)
	ss, exists := ps.Sectors[sector]
	if !exists {
		return
	}

	if amount > ps.Player.Resources.Budget {
		amount = ps.Player.Resources.Budget
	}
	if amount <= 0 {
		return
	}

	ps.Player.Resources.Budget -= amount
	ss.Investment += amount
	ss.Level += amount * InvestmentRate * ss.Growth
}

func (e *Engine) handlePolicy(ps *models.PlayerState, action models.Action) {
	// Data: {"policy": "increase_education_funding", "sector": "education"}
	sectorStr, ok := action.Data["sector"].(string)
	if !ok {
		return
	}
	sector := models.SectorType(sectorStr)
	ss, exists := ps.Sectors[sector]
	if !exists {
		return
	}

	// Policies cost influence but boost growth rate
	cost := 10.0
	if ps.Player.Resources.Influence < cost {
		return
	}
	ps.Player.Resources.Influence -= cost
	ss.Growth += 0.1
}

func (e *Engine) handleEventResponse(ps *models.PlayerState, action models.Action) {
	// Data: {"event_id": "xxx", "response": "mitigate", "sector": "security"}
	sectorStr, ok := action.Data["sector"].(string)
	if !ok {
		return
	}
	sector := models.SectorType(sectorStr)
	ss, exists := ps.Sectors[sector]
	if !exists {
		return
	}

	// Responding to events costs stability but reduces negative impact
	cost := 15.0
	if ps.Player.Resources.Stability < cost {
		return
	}
	ps.Player.Resources.Stability -= cost
	ss.Level += 5 // partial recovery
}

func (e *Engine) applyActiveEvents(g *models.Game) {
	for _, event := range g.ActiveEvents {
		for _, ps := range g.Players {
			// Skip if event is region-specific and not this player's region
			if event.RegionSpecific != "" && event.RegionSpecific != ps.Player.RegionID {
				continue
			}

			for sector, impact := range event.Impact {
				ss, exists := ps.Sectors[sector]
				if !exists {
					continue
				}
				ss.Level += impact
			}
		}
	}
}

func (e *Engine) applyDecay(ps *models.PlayerState) {
	for _, ss := range ps.Sectors {
		if ss.Investment == 0 {
			ss.Level -= DecayRate
		}
		// Reset investment for next turn
		ss.Investment = 0
	}
}

func (e *Engine) replenishResources(ps *models.PlayerState) {
	// Budget replenishes based on economics level
	econ := ps.Sectors[models.Economics]
	if econ != nil {
		ps.Player.Resources.Budget += 30 + (econ.Level * 0.5)
	}

	// Influence replenishes based on politics level
	pol := ps.Sectors[models.Politics]
	if pol != nil {
		ps.Player.Resources.Influence += 10 + (pol.Level * 0.2)
	}

	// Stability replenishes based on security level
	sec := ps.Sectors[models.Security]
	if sec != nil {
		ps.Player.Resources.Stability += 10 + (sec.Level * 0.2)
	}

	// Knowledge replenishes based on education + R&D
	edu := ps.Sectors[models.Education]
	rnd := ps.Sectors[models.RandD]
	if edu != nil && rnd != nil {
		ps.Player.Resources.Knowledge += 5 + (edu.Level+rnd.Level)*0.15
	}
}

func (e *Engine) calcScore(ps *models.PlayerState) float64 {
	total := 0.0
	for _, ss := range ps.Sectors {
		total += ss.Level
	}
	return total // max 500 (5 sectors * 100)
}

func (e *Engine) broadcastGameState(g *models.Game) {
	e.broadcast(g.ID, models.WSMessage{
		Type:    "game_state",
		Payload: g,
	})
}

func (e *Engine) broadcastEvents(g *models.Game, events []models.WorldEvent) {
	for _, evt := range events {
		e.broadcast(g.ID, models.WSMessage{
			Type:    "event",
			Payload: evt,
		})
	}
}

func (e *Engine) broadcastTurnResult(g *models.Game) {
	scores := make(map[string]float64)
	for pid, ps := range g.Players {
		scores[pid] = ps.Score
	}
	e.broadcast(g.ID, models.WSMessage{
		Type: "turn_result",
		Payload: map[string]interface{}{
			"turn":   g.CurrentTurn - 1,
			"scores": scores,
		},
	})
}
