package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rhaqim/worldgame/internal/external"
	"github.com/rhaqim/worldgame/internal/models"
)

type eventTemplate struct {
	Title           string
	Description     string
	Source          string
	AffectedSectors []models.SectorType
	Impact          map[models.SectorType]float64
	Duration        int
}

var eventPool = []eventTemplate{
	// Economic events
	{
		Title: "Global Market Crash", Description: "Stock markets tumble worldwide as investor confidence plummets.",
		Source: "economic_report", AffectedSectors: []models.SectorType{models.Economics, models.Politics},
		Impact: map[models.SectorType]float64{models.Economics: -15, models.Politics: -5}, Duration: 3,
	},
	{
		Title: "International Trade Deal", Description: "A landmark trade agreement opens new markets.",
		Source: "news", AffectedSectors: []models.SectorType{models.Economics, models.Politics},
		Impact: map[models.SectorType]float64{models.Economics: 10, models.Politics: 5}, Duration: 2,
	},
	{
		Title: "Commodity Boom", Description: "Surging demand for raw materials drives prices up.",
		Source: "economic_report", AffectedSectors: []models.SectorType{models.Economics},
		Impact: map[models.SectorType]float64{models.Economics: 12}, Duration: 2,
	},
	{
		Title: "Inflation Spike", Description: "Consumer prices rise sharply, eroding purchasing power.",
		Source: "economic_report", AffectedSectors: []models.SectorType{models.Economics, models.Education},
		Impact: map[models.SectorType]float64{models.Economics: -8, models.Education: -3}, Duration: 3,
	},
	{
		Title: "Tech IPO Frenzy", Description: "Multiple tech companies go public, boosting innovation funding.",
		Source: "news", AffectedSectors: []models.SectorType{models.Economics, models.RandD},
		Impact: map[models.SectorType]float64{models.Economics: 8, models.RandD: 6}, Duration: 1,
	},
	{
		Title: "Foreign Investment Surge", Description: "International investors flock to emerging opportunities.",
		Source: "economic_report", AffectedSectors: []models.SectorType{models.Economics},
		Impact: map[models.SectorType]float64{models.Economics: 10}, Duration: 2,
	},
	// Political events
	{
		Title: "Snap Election Called", Description: "Unexpected election creates political uncertainty.",
		Source: "news", AffectedSectors: []models.SectorType{models.Politics, models.Economics},
		Impact: map[models.SectorType]float64{models.Politics: -8, models.Economics: -4}, Duration: 2,
	},
	{
		Title: "Sweeping Policy Reform", Description: "Government enacts broad reforms across multiple sectors.",
		Source: "news", AffectedSectors: []models.SectorType{models.Politics, models.Education, models.Economics},
		Impact: map[models.SectorType]float64{models.Politics: 10, models.Education: 5, models.Economics: 3}, Duration: 3,
	},
	{
		Title: "Diplomatic Incident", Description: "A diplomatic blunder strains international relations.",
		Source: "news", AffectedSectors: []models.SectorType{models.Politics, models.Security},
		Impact: map[models.SectorType]float64{models.Politics: -10, models.Security: -5}, Duration: 2,
	},
	{
		Title: "Economic Sanctions Imposed", Description: "International community levies sanctions over policy disagreements.",
		Source: "news", AffectedSectors: []models.SectorType{models.Politics, models.Economics},
		Impact: map[models.SectorType]float64{models.Politics: -6, models.Economics: -12}, Duration: 4,
	},
	{
		Title: "International Summit Success", Description: "World leaders reach consensus on key issues at major summit.",
		Source: "news", AffectedSectors: []models.SectorType{models.Politics, models.Security},
		Impact: map[models.SectorType]float64{models.Politics: 12, models.Security: 5}, Duration: 2,
	},
	{
		Title: "Anti-Corruption Campaign", Description: "Government launches sweeping anti-corruption measures.",
		Source: "news", AffectedSectors: []models.SectorType{models.Politics},
		Impact: map[models.SectorType]float64{models.Politics: 8}, Duration: 3,
	},
	// Security events
	{
		Title: "Major Cyber Attack", Description: "Critical infrastructure targeted by sophisticated cyber operation.",
		Source: "news", AffectedSectors: []models.SectorType{models.Security, models.Economics, models.RandD},
		Impact: map[models.SectorType]float64{models.Security: -12, models.Economics: -5, models.RandD: -3}, Duration: 2,
	},
	{
		Title: "Border Dispute Escalates", Description: "Neighboring nations clash over territorial claims.",
		Source: "news", AffectedSectors: []models.SectorType{models.Security, models.Politics},
		Impact: map[models.SectorType]float64{models.Security: -10, models.Politics: -8}, Duration: 3,
	},
	{
		Title: "Terror Threat Level Raised", Description: "Intelligence agencies warn of imminent security threat.",
		Source: "news", AffectedSectors: []models.SectorType{models.Security, models.Economics},
		Impact: map[models.SectorType]float64{models.Security: -8, models.Economics: -4}, Duration: 2,
	},
	{
		Title: "Arms Deal Finalized", Description: "Major defense contract signed, boosting military capability.",
		Source: "news", AffectedSectors: []models.SectorType{models.Security, models.Economics},
		Impact: map[models.SectorType]float64{models.Security: 10, models.Economics: -3}, Duration: 2,
	},
	{
		Title: "Peacekeeping Mission Deployed", Description: "International peacekeeping force deployed to conflict zone.",
		Source: "news", AffectedSectors: []models.SectorType{models.Security, models.Politics},
		Impact: map[models.SectorType]float64{models.Security: 8, models.Politics: 6}, Duration: 3,
	},
	{
		Title: "Intelligence Leak", Description: "Classified documents leaked, exposing security vulnerabilities.",
		Source: "social_media", AffectedSectors: []models.SectorType{models.Security, models.Politics},
		Impact: map[models.SectorType]float64{models.Security: -10, models.Politics: -7}, Duration: 2,
	},
	// Education events
	{
		Title: "University Ranking Shakeup", Description: "Surprising shifts in global university rankings.",
		Source: "news", AffectedSectors: []models.SectorType{models.Education, models.RandD},
		Impact: map[models.SectorType]float64{models.Education: 6, models.RandD: 4}, Duration: 1,
	},
	{
		Title: "National Literacy Program", Description: "Government launches ambitious program to boost literacy rates.",
		Source: "news", AffectedSectors: []models.SectorType{models.Education},
		Impact: map[models.SectorType]float64{models.Education: 10}, Duration: 3,
	},
	{
		Title: "Brain Drain Crisis", Description: "Skilled workers and researchers emigrate in large numbers.",
		Source: "news", AffectedSectors: []models.SectorType{models.Education, models.RandD, models.Economics},
		Impact: map[models.SectorType]float64{models.Education: -8, models.RandD: -10, models.Economics: -4}, Duration: 3,
	},
	{
		Title: "International Scholarship Fund", Description: "Massive scholarship program attracts global talent.",
		Source: "news", AffectedSectors: []models.SectorType{models.Education, models.RandD},
		Impact: map[models.SectorType]float64{models.Education: 8, models.RandD: 5}, Duration: 2,
	},
	{
		Title: "Education Workers Strike", Description: "Teachers and professors walk out demanding better conditions.",
		Source: "social_media", AffectedSectors: []models.SectorType{models.Education, models.Politics},
		Impact: map[models.SectorType]float64{models.Education: -10, models.Politics: -4}, Duration: 2,
	},
	{
		Title: "Digital Learning Revolution", Description: "New e-learning platform transforms access to education.",
		Source: "news", AffectedSectors: []models.SectorType{models.Education, models.RandD},
		Impact: map[models.SectorType]float64{models.Education: 7, models.RandD: 3}, Duration: 2,
	},
	// R&D events
	{
		Title: "Breakthrough Discovery", Description: "Scientists announce a paradigm-shifting discovery.",
		Source: "news", AffectedSectors: []models.SectorType{models.RandD, models.Education},
		Impact: map[models.SectorType]float64{models.RandD: 15, models.Education: 5}, Duration: 2,
	},
	{
		Title: "Patent War Erupts", Description: "Major corporations battle over key technology patents.",
		Source: "news", AffectedSectors: []models.SectorType{models.RandD, models.Economics},
		Impact: map[models.SectorType]float64{models.RandD: -5, models.Economics: -3}, Duration: 2,
	},
	{
		Title: "Space Mission Launches", Description: "Ambitious space mission captures global attention.",
		Source: "news", AffectedSectors: []models.SectorType{models.RandD, models.Education},
		Impact: map[models.SectorType]float64{models.RandD: 12, models.Education: 4}, Duration: 2,
	},
	{
		Title: "Pandemic Research Accelerates", Description: "New funding poured into disease prevention research.",
		Source: "news", AffectedSectors: []models.SectorType{models.RandD, models.Security},
		Impact: map[models.SectorType]float64{models.RandD: 10, models.Security: 3}, Duration: 3,
	},
	{
		Title: "AI Regulation Debate", Description: "Governments struggle to regulate rapidly advancing AI technology.",
		Source: "social_media", AffectedSectors: []models.SectorType{models.RandD, models.Politics, models.Economics},
		Impact: map[models.SectorType]float64{models.RandD: -4, models.Politics: -3, models.Economics: 2}, Duration: 2,
	},
	{
		Title: "Green Energy Breakthrough", Description: "New renewable energy tech promises to slash carbon emissions.",
		Source: "news", AffectedSectors: []models.SectorType{models.RandD, models.Economics},
		Impact: map[models.SectorType]float64{models.RandD: 10, models.Economics: 6}, Duration: 2,
	},
	{
		Title: "Research Lab Explosion", Description: "Accident at major research facility halts critical projects.",
		Source: "news", AffectedSectors: []models.SectorType{models.RandD, models.Security},
		Impact: map[models.SectorType]float64{models.RandD: -12, models.Security: -4}, Duration: 2,
	},
}

type EventGenerator struct {
	rng   *rand.Rand
	feeds *external.FeedProvider
}

func NewEventGenerator(feeds *external.FeedProvider) *EventGenerator {
	return &EventGenerator{
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
		feeds: feeds,
	}
}

func (eg *EventGenerator) GenerateEvents(gameState *models.Game, count int) []models.WorldEvent {
	events := make([]models.WorldEvent, 0, count)

	// Get feed context to bias event selection
	ctx := eg.feeds.GenerateEventContext()
	_ = ctx // used for future weighting; for now pick randomly

	usedIndices := map[int]bool{}
	playerIDs := make([]string, 0, len(gameState.Players))
	for pid := range gameState.Players {
		playerIDs = append(playerIDs, pid)
	}

	for i := 0; i < count; i++ {
		// Pick a random template not yet used this round
		idx := eg.rng.Intn(len(eventPool))
		for usedIndices[idx] {
			idx = eg.rng.Intn(len(eventPool))
		}
		usedIndices[idx] = true
		tmpl := eventPool[idx]

		// Decide if event is global or region-specific (30% chance region-specific)
		regionSpecific := ""
		if eg.rng.Float64() < 0.3 && len(playerIDs) > 0 {
			ps := gameState.Players[playerIDs[eg.rng.Intn(len(playerIDs))]]
			regionSpecific = ps.Player.RegionID
		}

		// Apply slight randomness to impact values
		impact := make(map[models.SectorType]float64)
		for s, v := range tmpl.Impact {
			noise := (eg.rng.Float64() - 0.5) * 4 // +/- 2
			impact[s] = v + noise
		}

		event := models.WorldEvent{
			ID:              fmt.Sprintf("evt_%d_%d_%d", gameState.CurrentTurn, i, eg.rng.Intn(10000)),
			Title:           tmpl.Title,
			Description:     tmpl.Description,
			Source:          tmpl.Source,
			AffectedSectors: tmpl.AffectedSectors,
			Impact:          impact,
			Duration:        tmpl.Duration,
			RegionSpecific:  regionSpecific,
		}
		events = append(events, event)
	}

	return events
}
