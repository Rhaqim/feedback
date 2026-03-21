package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
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

// ChallengeGenerator creates challenges for a game based on its tags and region.
type ChallengeGenerator struct {
	rng *rand.Rand
}

// NewChallengeGenerator creates a new ChallengeGenerator.
func NewChallengeGenerator() *ChallengeGenerator {
	return &ChallengeGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateChallenges creates challenges for a game. It produces countPerTag
// challenges for each of the game's active tags, customized with the game's
// region name.
func (cg *ChallengeGenerator) GenerateChallenges(g *models.Game, countPerTag int) []models.Challenge {
	challenges := make([]models.Challenge, 0)

	for _, tag := range g.Tags {
		templates, ok := challengePool[tag]
		if !ok || len(templates) == 0 {
			continue
		}

		// Shuffle and pick up to countPerTag unique templates
		perm := cg.rng.Perm(len(templates))
		count := countPerTag
		if count > len(templates) {
			count = len(templates)
		}

		for i := 0; i < count; i++ {
			tmpl := templates[perm[i]]
			severity := tmpl.MinSeverity + cg.rng.Intn(tmpl.MaxSeverity-tmpl.MinSeverity+1)

			title := tmpl.TitleFmt
			desc := tmpl.DescFmt
			if strings.Contains(title, "%s") {
				title = fmt.Sprintf(tmpl.TitleFmt, g.RegionName)
			}
			if strings.Contains(desc, "%s") {
				desc = fmt.Sprintf(tmpl.DescFmt, g.RegionName)
			}

			challenge := models.Challenge{
				ID:          fmt.Sprintf("ch_%s_w%d_%s_%d", g.ID, g.WeekNumber, tag, i),
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
	}

	return challenges
}
