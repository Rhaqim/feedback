package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
)

// tagKeywords maps each tag to terms that a good proposal would reference.
var tagKeywords = map[models.Tag][]string{
	models.TagSecurity:  {"firewall", "encryption", "surveillance", "defense", "patrol", "intelligence", "audit", "training", "protocol", "monitoring", "threat", "response", "prevention", "cyber", "enforcement"},
	models.TagPolitics:  {"transparency", "accountability", "reform", "legislation", "coalition", "governance", "participation", "oversight", "democracy", "policy", "regulation", "mandate", "consensus", "diplomacy", "institution"},
	models.TagEconomics: {"investment", "subsidy", "tariff", "incentive", "diversification", "employment", "fiscal", "monetary", "trade", "budget", "growth", "inflation", "market", "taxation", "enterprise"},
	models.TagEducation: {"curriculum", "scholarship", "teacher", "digital", "literacy", "enrollment", "research", "university", "training", "access", "funding", "assessment", "inclusion", "mentorship", "innovation"},
	models.TagLaw:       {"legislation", "enforcement", "judiciary", "regulation", "compliance", "rights", "amendment", "tribunal", "arbitration", "prosecution", "statute", "precedent", "due process", "reform", "oversight"},
}

// feedbackTemplates provide short AI feedback messages.
var feedbackTemplates = []string{
	"Solid approach with clear actionable steps.",
	"Good strategic thinking but could use more implementation detail.",
	"Creative solution that addresses root causes effectively.",
	"Well-structured proposal with realistic resource allocation.",
	"Innovative idea, though feasibility needs further consideration.",
	"Comprehensive plan that accounts for multiple stakeholder perspectives.",
	"Promising direction but lacks specificity in key areas.",
	"Strong analytical foundation with practical recommendations.",
	"Thoughtful response that balances short-term and long-term impact.",
	"Effective proposal that leverages existing institutional frameworks.",
}

// Evaluator scores proposals using a mock AI evaluation system.
type Evaluator struct {
	rng *rand.Rand
}

// NewEvaluator creates a new Evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// EvaluateProposals scores every proposal in the game and determines the
// weekly winner. It modifies the proposals in place, sets AIScore and
// AIFeedback on each, and returns the WeekWinner.
func (ev *Evaluator) EvaluateProposals(g *models.Game) *models.WeekWinner {
	if len(g.Proposals) == 0 {
		return nil
	}

	// Build a challenge-tag lookup so we can match proposals to tags.
	challengeTagMap := make(map[string]models.Tag)
	for _, ch := range g.Challenges {
		challengeTagMap[ch.ID] = ch.Tag
	}

	// Score each proposal.
	for i := range g.Proposals {
		p := &g.Proposals[i]
		tag := challengeTagMap[p.ChallengeID]
		p.AIScore = ev.scoreProposal(p, tag)
		p.AIFeedback = feedbackTemplates[ev.rng.Intn(len(feedbackTemplates))]
	}

	// Aggregate scores per player.
	// Winner = player with highest sum of (ai_score * points_invested_weight).
	type playerAgg struct {
		totalWeightedScore float64
		playerID           string
		playerName         string
	}

	agg := make(map[string]*playerAgg)
	for _, p := range g.Proposals {
		a, ok := agg[p.PlayerID]
		if !ok {
			a = &playerAgg{playerID: p.PlayerID, playerName: p.PlayerName}
			agg[p.PlayerID] = a
		}
		// Weight: points invested normalised so that 100 points = 1.0x multiplier
		weight := 1.0 + (p.PointsInvested / 100.0)
		a.totalWeightedScore += p.AIScore * weight
	}

	var best *playerAgg
	for _, a := range agg {
		if best == nil || a.totalWeightedScore > best.totalWeightedScore {
			best = a
		}
	}

	if best == nil {
		return nil
	}

	return &models.WeekWinner{
		PlayerID:   best.playerID,
		PlayerName: best.playerName,
		Score:      math.Round(best.totalWeightedScore*100) / 100,
		Summary:    fmt.Sprintf("%s achieved the highest combined proposal score of %.1f across all submissions.", best.playerName, best.totalWeightedScore),
	}
}

// scoreProposal computes a 0-100 score for a single proposal.
func (ev *Evaluator) scoreProposal(p *models.Proposal, tag models.Tag) float64 {
	// 1. Length / detail score (0-40 points).
	// Longer, more detailed descriptions score higher up to a ceiling.
	wordCount := float64(len(strings.Fields(p.Description)))
	lengthScore := math.Min(40, wordCount*0.8) // ~50 words for full marks

	// 2. Keyword relevance score (0-30 points).
	keywords := tagKeywords[tag]
	lowerDesc := strings.ToLower(p.Description)
	hits := 0
	for _, kw := range keywords {
		if strings.Contains(lowerDesc, kw) {
			hits++
		}
	}
	keywordScore := math.Min(30, float64(hits)*6) // 5 keywords for full marks

	// 3. Points invested bonus (0-20 points).
	// Higher investment shows conviction.
	investmentScore := math.Min(20, p.PointsInvested*0.2)

	// 4. Random variance (-5 to +10 points) for variety.
	randomBonus := (ev.rng.Float64() * 15) - 5

	total := lengthScore + keywordScore + investmentScore + randomBonus
	// Clamp to 0-100.
	total = math.Max(0, math.Min(100, total))
	return math.Round(total*10) / 10
}
