package game

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
	"github.com/rhaqim/worldgame/internal/store"
)

// challengeTemplate holds a reusable challenge skeleton.
type challengeTemplate struct {
	TitleFmt    string // use %s for region name substitution
	DescFmt     string // use %s for region name substitution
	Source      string // "news", "social_media", "report"
	MinSeverity int
	MaxSeverity int
}

// challengePool maps each tag to a slice of templates.
var challengePool = map[models.Tag][]challengeTemplate{
	models.TagSecurity: {
		{
			TitleFmt: "Cybercrime Wave Hits %s", DescFmt: "A surge in ransomware and phishing attacks targets critical infrastructure and financial institutions across %s, overwhelming existing cyber defense capabilities.",
			Source: "news", MinSeverity: 6, MaxSeverity: 9,
		},
		{
			TitleFmt: "Border Security Crisis in %s", DescFmt: "Porous borders in %s have led to increased smuggling and unauthorized crossings, straining law enforcement and diplomatic relations with neighboring states.",
			Source: "report", MinSeverity: 5, MaxSeverity: 8,
		},
		{
			TitleFmt: "Infrastructure Vulnerability Exposed in %s", DescFmt: "Security audits reveal that power grids, water systems, and transportation networks in %s are critically vulnerable to both physical and cyber attacks.",
			Source: "report", MinSeverity: 7, MaxSeverity: 10,
		},
		{
			TitleFmt: "Data Privacy Breach Scandal in %s", DescFmt: "Millions of citizens in %s have their personal data exposed after a massive breach at a government database, raising questions about data protection standards.",
			Source: "news", MinSeverity: 5, MaxSeverity: 8,
		},
		{
			TitleFmt: "Extremist Threat Escalation in %s", DescFmt: "Intelligence agencies in %s warn of growing radicalization networks operating both online and in rural areas, calling for new counter-terrorism strategies.",
			Source: "social_media", MinSeverity: 7, MaxSeverity: 10,
		},
	},
	models.TagPolitics: {
		{
			TitleFmt: "Corruption Scandal Rocks %s", DescFmt: "Senior government officials in %s are implicated in a wide-reaching corruption scandal involving public procurement contracts worth billions.",
			Source: "news", MinSeverity: 6, MaxSeverity: 9,
		},
		{
			TitleFmt: "Election Integrity Under Fire in %s", DescFmt: "Allegations of voter suppression, misinformation campaigns, and foreign interference cast doubt on upcoming elections in %s.",
			Source: "social_media", MinSeverity: 7, MaxSeverity: 10,
		},
		{
			TitleFmt: "Policy Gridlock Paralyzes %s", DescFmt: "Deep partisan divisions in %s have brought legislative progress to a halt, with critical bills on healthcare, infrastructure, and climate stalled indefinitely.",
			Source: "news", MinSeverity: 4, MaxSeverity: 7,
		},
		{
			TitleFmt: "Diplomatic Tensions Flare for %s", DescFmt: "Deteriorating diplomatic relations between %s and key trading partners threaten regional stability and economic cooperation agreements.",
			Source: "news", MinSeverity: 5, MaxSeverity: 8,
		},
		{
			TitleFmt: "Governance Reform Demanded in %s", DescFmt: "Mass public protests in %s demand sweeping governance reforms including greater transparency, decentralization, and anti-corruption measures.",
			Source: "social_media", MinSeverity: 5, MaxSeverity: 8,
		},
	},
	models.TagEconomics: {
		{
			TitleFmt: "Inflation Crisis Grips %s", DescFmt: "Consumer prices in %s have surged to multi-year highs, eroding purchasing power and threatening to push millions into poverty as wages fail to keep pace.",
			Source: "report", MinSeverity: 6, MaxSeverity: 9,
		},
		{
			TitleFmt: "Unemployment Surge in %s", DescFmt: "Factory closures and automation have driven unemployment in %s to alarming levels, particularly among youth and unskilled workers in urban centers.",
			Source: "news", MinSeverity: 6, MaxSeverity: 9,
		},
		{
			TitleFmt: "Trade Deficit Widens for %s", DescFmt: "Imports continue to outpace exports in %s as global demand shifts, currency pressures mount, and key industries lose competitive advantage.",
			Source: "report", MinSeverity: 4, MaxSeverity: 7,
		},
		{
			TitleFmt: "Housing Market Crisis in %s", DescFmt: "Skyrocketing housing costs in %s have made homeownership unattainable for a generation, while speculative investment and short-term rentals deplete available stock.",
			Source: "news", MinSeverity: 5, MaxSeverity: 8,
		},
		{
			TitleFmt: "Startup Ecosystem Struggles in %s", DescFmt: "Despite early promise, the startup ecosystem in %s faces a funding drought as venture capital retreats, regulatory burdens increase, and talent emigrates.",
			Source: "social_media", MinSeverity: 3, MaxSeverity: 6,
		},
	},
	models.TagEducation: {
		{
			TitleFmt: "Literacy Gap Widens in %s", DescFmt: "Rural and underserved communities in %s fall further behind in literacy rates, with overcrowded classrooms and insufficient teaching resources compounding the problem.",
			Source: "report", MinSeverity: 5, MaxSeverity: 8,
		},
		{
			TitleFmt: "Digital Divide Deepens in %s", DescFmt: "Millions of students in %s lack reliable internet access and devices for digital learning, creating a two-tier education system that entrenches inequality.",
			Source: "news", MinSeverity: 5, MaxSeverity: 8,
		},
		{
			TitleFmt: "Brain Drain Accelerates from %s", DescFmt: "Top graduates and researchers are leaving %s in record numbers for better opportunities abroad, depleting the nation's intellectual capital and innovation capacity.",
			Source: "news", MinSeverity: 6, MaxSeverity: 9,
		},
		{
			TitleFmt: "Curriculum Reform Debate in %s", DescFmt: "Educators, parents, and policymakers in %s clash over proposed curriculum changes designed to better prepare students for a technology-driven economy.",
			Source: "social_media", MinSeverity: 3, MaxSeverity: 6,
		},
		{
			TitleFmt: "Research Funding Crisis in %s", DescFmt: "Universities and research institutions in %s face severe budget cuts, threatening ongoing scientific projects and the country's ability to compete in global innovation.",
			Source: "report", MinSeverity: 5, MaxSeverity: 8,
		},
	},
	models.TagLaw: {
		{
			TitleFmt: "Judicial Backlog Overwhelms %s", DescFmt: "Courts in %s are buried under a massive case backlog, with some cases taking years to reach trial. Justice delayed is justice denied for millions of citizens.",
			Source: "report", MinSeverity: 5, MaxSeverity: 8,
		},
		{
			TitleFmt: "Regulatory Gaps Exposed in %s", DescFmt: "Rapid technological change has outpaced the legal framework in %s, leaving cryptocurrency, AI, and gig economy workers in a regulatory gray zone.",
			Source: "news", MinSeverity: 4, MaxSeverity: 7,
		},
		{
			TitleFmt: "Human Rights Enforcement Challenged in %s", DescFmt: "International observers raise concerns about human rights enforcement in %s, citing issues with press freedom, detention conditions, and minority protections.",
			Source: "report", MinSeverity: 7, MaxSeverity: 10,
		},
		{
			TitleFmt: "Intellectual Property Disputes Surge in %s", DescFmt: "A wave of IP theft and patent infringement cases in %s threatens foreign investment and undermines the domestic innovation ecosystem.",
			Source: "news", MinSeverity: 4, MaxSeverity: 7,
		},
		{
			TitleFmt: "Constitutional Reform Push in %s", DescFmt: "Growing calls for constitutional amendments in %s seek to address outdated provisions on term limits, judicial independence, and citizens' digital rights.",
			Source: "social_media", MinSeverity: 5, MaxSeverity: 8,
		},
	},
}

// SeedTemplate is a challenge template entry exposed for database seeding.
type SeedTemplate struct {
	Title       string
	Description string
	Source      string
}

// GetChallengePoolForSeeding returns all hardcoded challenge templates
// organized by tag, suitable for inserting into the database as seed data.
func GetChallengePoolForSeeding() map[models.Tag][]SeedTemplate {
	result := make(map[models.Tag][]SeedTemplate)
	for tag, templates := range challengePool {
		seeds := make([]SeedTemplate, len(templates))
		for i, t := range templates {
			seeds[i] = SeedTemplate{
				Title:       t.TitleFmt,
				Description: t.DescFmt,
				Source:      t.Source,
			}
		}
		result[tag] = seeds
	}
	return result
}

// ChallengeGenerator creates challenges for a game based on its tags and region.
type ChallengeGenerator struct {
	rng   *rand.Rand
	store *store.Store
}

// NewChallengeGenerator creates a new ChallengeGenerator backed by the store.
func NewChallengeGenerator(s *store.Store) *ChallengeGenerator {
	return &ChallengeGenerator{
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
		store: s,
	}
}

// GenerateChallenges creates challenges for a game using a three-tier priority:
// 1. Curated challenges from the Game Master (highest priority)
// 2. Real feed items from the database
// 3. Static challenge templates (fallback)
func (cg *ChallengeGenerator) GenerateChallenges(ctx context.Context, g *models.Game, countPerTag int) []models.Challenge {
	challenges := make([]models.Challenge, 0)

	for _, tag := range g.Tags {
		// Priority 1: Curated challenges from Game Master.
		curatedChallenges := cg.generateFromCurated(ctx, g, tag, countPerTag)
		challenges = append(challenges, curatedChallenges...)

		remaining := countPerTag - len(curatedChallenges)
		if remaining <= 0 {
			continue
		}

		// Priority 2: Raw feed items.
		feedChallenges := cg.generateFromFeeds(ctx, g, tag, remaining)
		challenges = append(challenges, feedChallenges...)

		remaining -= len(feedChallenges)
		if remaining > 0 {
			// Priority 3: Static templates (fallback).
			offset := len(curatedChallenges) + len(feedChallenges)
			templateChallenges := cg.generateFromTemplates(ctx, g, tag, remaining, offset)
			challenges = append(challenges, templateChallenges...)
		}
	}

	return challenges
}

// generateFromCurated creates challenges from curated items approved by the Game Master.
func (cg *ChallengeGenerator) generateFromCurated(ctx context.Context, g *models.Game, tag models.Tag, count int) []models.Challenge {
	curatedItems, err := cg.store.GetUnusedCuratedChallenges(ctx, tag, g.RegionID, count)
	if err != nil {
		log.Printf("[ChallengeGen] Error loading curated challenges for tag %s: %v", tag, err)
		return nil
	}

	var challenges []models.Challenge
	for i, cc := range curatedItems {
		challenge := models.Challenge{
			ID:          fmt.Sprintf("ch_%s_w%d_%s_%d", g.ID, g.WeekNumber, tag, i),
			Tag:         tag,
			Title:       cc.Title,
			Description: cc.Description,
			Source:      cc.Source,
			Region:      g.RegionName,
			Severity:    cc.Severity,
			CreatedAt:   time.Now(),
			Active:      true,
		}
		challenges = append(challenges, challenge)

		if err := cg.store.MarkCuratedChallengeUsed(ctx, cc.ID); err != nil {
			log.Printf("[ChallengeGen] Error marking curated challenge %d as used: %v", cc.ID, err)
		}
	}

	if len(challenges) > 0 {
		log.Printf("[ChallengeGen] Generated %d curated challenges for tag %s", len(challenges), tag)
	}
	return challenges
}

// generateFromFeeds creates challenges from real feed items stored in the database.
func (cg *ChallengeGenerator) generateFromFeeds(ctx context.Context, g *models.Game, tag models.Tag, count int) []models.Challenge {
	feedItems, err := cg.store.GetUnusedFeedItems(ctx, tag, g.RegionID, count)
	if err != nil {
		log.Printf("[ChallengeGen] Error loading feed items for tag %s: %v", tag, err)
		return nil
	}

	var challenges []models.Challenge
	for i, fi := range feedItems {
		severity := cg.calculateSeverity(fi.Title, fi.Description)

		challenge := models.Challenge{
			ID:          fmt.Sprintf("ch_%s_w%d_%s_%d", g.ID, g.WeekNumber, tag, i),
			Tag:         tag,
			Title:       fi.Title,
			Description: fi.Description,
			Source:      fi.FeedName,
			Region:      g.RegionName,
			Severity:    severity,
			CreatedAt:   time.Now(),
			Active:      true,
		}
		challenges = append(challenges, challenge)

		// Mark the feed item as used.
		if err := cg.store.MarkFeedItemUsed(ctx, fi.ID); err != nil {
			log.Printf("[ChallengeGen] Error marking feed item %d as used: %v", fi.ID, err)
		}
	}

	if len(challenges) > 0 {
		log.Printf("[ChallengeGen] Generated %d feed-based challenges for tag %s", len(challenges), tag)
	}
	return challenges
}

// generateFromTemplates creates challenges from static templates (fallback).
func (cg *ChallengeGenerator) generateFromTemplates(ctx context.Context, g *models.Game, tag models.Tag, count int, offset int) []models.Challenge {
	templates, err := cg.store.GetChallengeTemplates(ctx, tag)
	if err != nil {
		log.Printf("[ChallengeGen] Error loading templates for tag %s: %v", tag, err)
		return nil
	}
	if len(templates) == 0 {
		return nil
	}

	// Shuffle and pick up to count unique templates.
	perm := cg.rng.Perm(len(templates))
	if count > len(templates) {
		count = len(templates)
	}

	var challenges []models.Challenge
	for i := 0; i < count; i++ {
		tmpl := templates[perm[i]]
		severity := 5 + cg.rng.Intn(6) // 5-10 range

		title := tmpl.TitleTemplate
		desc := tmpl.DescriptionTemplate
		if strings.Contains(title, "%s") {
			title = fmt.Sprintf(tmpl.TitleTemplate, g.RegionName)
		}
		if strings.Contains(desc, "%s") {
			desc = fmt.Sprintf(tmpl.DescriptionTemplate, g.RegionName)
		}

		challenge := models.Challenge{
			ID:          fmt.Sprintf("ch_%s_w%d_%s_%d", g.ID, g.WeekNumber, tag, offset+i),
			Tag:         tag,
			Title:       title,
			Description: desc,
			Source:      tmpl.Source,
			Region:      g.RegionName,
			Severity:    severity,
			CreatedAt:   time.Now(),
			Active:      true,
		}
		challenges = append(challenges, challenge)
	}

	if len(challenges) > 0 {
		log.Printf("[ChallengeGen] Generated %d template-based challenges for tag %s (fallback)", len(challenges), tag)
	}
	return challenges
}

// severityKeywords maps keywords to severity weight modifiers.
var severityKeywords = map[string]int{
	"crisis":    3,
	"collapse":  3,
	"war":       3,
	"attack":    2,
	"threat":    2,
	"surge":     2,
	"critical":  2,
	"emergency": 3,
	"breach":    2,
	"scandal":   2,
	"crash":     2,
	"shortage":  1,
	"risk":      1,
	"concern":   1,
	"decline":   1,
	"reform":    0,
	"growth":    -1,
	"recovery":  -1,
	"progress":  -1,
}

// calculateSeverity estimates severity (1-10) based on feed item content.
func (cg *ChallengeGenerator) calculateSeverity(title, description string) int {
	text := strings.ToLower(title + " " + description)
	score := 5 // baseline

	for keyword, weight := range severityKeywords {
		if strings.Contains(text, keyword) {
			score += weight
		}
	}

	// Clamp to 1-10.
	if score < 1 {
		score = 1
	}
	if score > 10 {
		score = 10
	}

	// Add a little randomness.
	score += cg.rng.Intn(3) - 1 // -1, 0, or +1
	if score < 1 {
		score = 1
	}
	if score > 10 {
		score = 10
	}

	return score
}
