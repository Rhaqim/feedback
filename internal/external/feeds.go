package external

import (
	"math/rand"
	"time"
)

// FeedItem represents a single item from an external feed source.
type FeedItem struct {
	Headline  string  `json:"headline"`
	Source    string  `json:"source"`    // "news", "social_media", "report"
	Sentiment float64 `json:"sentiment"` // -1.0 to 1.0
	Tag       string  `json:"tag"`
	Region    string  `json:"region"`
}

// tagHeadlines provides mock feed data per tag that can be used
// to contextualise challenge generation.
var tagHeadlines = map[string][]struct {
	headline  string
	source    string
	sentiment float64
}{
	"security": {
		{"Ransomware Attack Cripples Hospital Network", "news", -0.8},
		{"Border Patrol Intercepts Record Smuggling Haul", "news", 0.3},
		{"Data Breach Exposes Millions of Citizen Records", "news", -0.9},
		{"New Cybersecurity Task Force Announced", "report", 0.6},
		{"Extremist Propaganda Spreads on Social Platforms", "social_media", -0.7},
		{"Critical Infrastructure Audit Reveals Gaps", "report", -0.5},
		{"Intelligence Sharing Agreement Signed", "news", 0.7},
		{"Drone Surveillance Raises Privacy Concerns", "social_media", -0.3},
	},
	"politics": {
		{"Corruption Probe Targets Senior Officials", "news", -0.7},
		{"Youth Voter Turnout Hits Record Highs", "social_media", 0.5},
		{"Opposition Calls for Snap Election", "news", -0.4},
		{"Government Transparency Report Released", "report", 0.3},
		{"Diplomatic Rift Deepens Over Trade Policy", "news", -0.6},
		{"Constitutional Amendment Proposed", "news", 0.2},
		{"Grassroots Movement Gains Momentum", "social_media", 0.6},
		{"Policy Gridlock Stalls Critical Legislation", "report", -0.5},
	},
	"economics": {
		{"Inflation Reaches Multi-Year High", "report", -0.7},
		{"Central Bank Signals Rate Hike", "news", -0.3},
		{"Startup Funding Round Breaks Records", "news", 0.8},
		{"Unemployment Claims Surge Amid Layoffs", "report", -0.8},
		{"Trade Surplus Narrows as Imports Rise", "report", -0.4},
		{"Housing Prices Soar Beyond Reach", "social_media", -0.6},
		{"Manufacturing Sector Shows Recovery Signs", "news", 0.5},
		{"Foreign Direct Investment Flows Increase", "report", 0.6},
	},
	"education": {
		{"Schools Struggle With Teacher Shortages", "news", -0.6},
		{"Digital Learning Platform Launches Nationwide", "news", 0.7},
		{"University Rankings Show Surprising Shifts", "report", 0.3},
		{"Student Loan Debt Crisis Deepens", "social_media", -0.7},
		{"STEM Scholarship Program Expands", "news", 0.6},
		{"Rural Schools Face Connectivity Gap", "report", -0.5},
		{"Brain Drain Accelerates Among Graduates", "news", -0.8},
		{"Research Funding Cuts Threaten Innovation", "report", -0.6},
	},
	"law": {
		{"Court Backlog Leaves Thousands Waiting for Justice", "report", -0.7},
		{"New Data Privacy Legislation Introduced", "news", 0.5},
		{"Human Rights Report Criticises Detention Conditions", "report", -0.8},
		{"IP Theft Cases Surge in Tech Sector", "news", -0.5},
		{"Judicial Reform Bill Gains Bipartisan Support", "news", 0.6},
		{"Regulatory Framework for AI Proposed", "report", 0.4},
		{"Legal Aid Funding Slashed", "social_media", -0.6},
		{"Constitutional Court Rules on Digital Rights", "news", 0.3},
	},
}

// FeedProvider generates mock external feed data for contextualising
// challenge generation.
type FeedProvider struct {
	rng *rand.Rand
}

// NewFeedProvider creates a new FeedProvider.
func NewFeedProvider() *FeedProvider {
	return &FeedProvider{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetFeedItems returns mock feed items for a specific tag and region.
func (f *FeedProvider) GetFeedItems(tag string, region string, count int) []FeedItem {
	headlines, ok := tagHeadlines[tag]
	if !ok {
		return nil
	}

	items := make([]FeedItem, 0, count)
	for i := 0; i < count; i++ {
		idx := f.rng.Intn(len(headlines))
		h := headlines[idx]
		items = append(items, FeedItem{
			Headline:  h.headline,
			Source:    h.source,
			Sentiment: h.sentiment + (f.rng.Float64()-0.5)*0.2,
			Tag:       tag,
			Region:    region,
		})
	}
	return items
}

// GetFeedContext returns aggregated sentiment data across all tags for a
// region. This can be used by challenge generators to weight severity.
func (f *FeedProvider) GetFeedContext(region string) map[string]float64 {
	context := make(map[string]float64)
	for tag := range tagHeadlines {
		items := f.GetFeedItems(tag, region, 3)
		avgSentiment := 0.0
		for _, item := range items {
			avgSentiment += item.Sentiment
		}
		if len(items) > 0 {
			avgSentiment /= float64(len(items))
		}
		context[tag+"_sentiment"] = avgSentiment
	}
	return context
}
