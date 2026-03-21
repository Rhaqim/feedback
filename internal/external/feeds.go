package external

import (
	"fmt"
	"math/rand"
	"time"
)

type NewsItem struct {
	Headline  string  `json:"headline"`
	Source    string  `json:"source"`
	Sentiment float64 `json:"sentiment"` // -1.0 to 1.0
	Category string  `json:"category"`
}

type SocialTrend struct {
	Topic      string  `json:"topic"`
	Engagement float64 `json:"engagement"` // 0-100
	Sentiment  float64 `json:"sentiment"`
	Region     string  `json:"region"`
}

type EconomicIndicator struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Change float64 `json:"change"` // percentage change
	Region string  `json:"region"`
}

var newsHeadlines = []struct {
	headline  string
	category  string
	sentiment float64
}{
	{"Global Markets Rally on Trade Optimism", "economics", 0.7},
	{"Tensions Rise Over Disputed Territory", "security", -0.6},
	{"Breakthrough in Quantum Computing Announced", "randd", 0.9},
	{"Education Reform Bill Passes Parliament", "education", 0.5},
	{"Opposition Party Gains Ground in Polls", "politics", -0.2},
	{"Oil Prices Surge After Supply Disruption", "economics", -0.4},
	{"New Cybersecurity Threat Targets Infrastructure", "security", -0.8},
	{"University Opens Largest AI Research Lab", "randd", 0.8},
	{"Literacy Rates Hit All-Time High", "education", 0.9},
	{"Diplomatic Summit Yields Historic Agreement", "politics", 0.8},
	{"Central Bank Raises Interest Rates", "economics", -0.3},
	{"Peacekeeping Forces Deploy to Conflict Zone", "security", 0.3},
	{"Space Agency Announces Mars Mission Timeline", "randd", 0.7},
	{"Student Protests Demand Curriculum Changes", "education", -0.2},
	{"Corruption Scandal Rocks Ruling Party", "politics", -0.9},
	{"Tech Giant Reports Record Quarterly Earnings", "economics", 0.6},
	{"Border Patrol Intercepts Smuggling Network", "security", 0.4},
	{"Clinical Trials Show Promise for New Vaccine", "randd", 0.8},
	{"Teacher Shortage Reaches Critical Levels", "education", -0.7},
	{"International Sanctions Imposed on Regime", "politics", -0.5},
	{"Cryptocurrency Market Experiences Flash Crash", "economics", -0.6},
	{"Cyber Attack Disrupts National Power Grid", "security", -0.9},
	{"Renewable Energy Breakthrough Cuts Costs by 40%", "randd", 0.9},
	{"STEM Scholarship Program Expands Nationwide", "education", 0.7},
	{"Coalition Government Formed After Deadlock", "politics", 0.3},
}

var trendingTopics = []struct {
	topic    string
	category string
}{
	{"#ClimateAction", "politics"},
	{"#TechInnovation", "randd"},
	{"#EconomicRecovery", "economics"},
	{"#EducationMatters", "education"},
	{"#CyberSecurity", "security"},
	{"#TradeWar", "economics"},
	{"#SpaceExploration", "randd"},
	{"#ElectionDay", "politics"},
	{"#OnlineLearning", "education"},
	{"#MilitarySpending", "security"},
}

var economicIndicators = []string{
	"GDP Growth Rate",
	"Unemployment Rate",
	"Inflation Rate",
	"Trade Balance",
	"Consumer Confidence Index",
	"Manufacturing PMI",
	"Foreign Direct Investment",
	"Stock Market Index",
}

type FeedProvider struct {
	rng *rand.Rand
}

func NewFeedProvider() *FeedProvider {
	return &FeedProvider{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (f *FeedProvider) GetNewsItems(count int) []NewsItem {
	items := make([]NewsItem, 0, count)
	sources := []string{"World Times", "Global Herald", "The Observer", "News Network", "Daily Report"}

	for i := 0; i < count; i++ {
		idx := f.rng.Intn(len(newsHeadlines))
		h := newsHeadlines[idx]
		items = append(items, NewsItem{
			Headline:  h.headline,
			Source:    sources[f.rng.Intn(len(sources))],
			Sentiment: h.sentiment + (f.rng.Float64()-0.5)*0.2, // add noise
			Category:  h.category,
		})
	}
	return items
}

func (f *FeedProvider) GetSocialTrends(count int, regionID string) []SocialTrend {
	trends := make([]SocialTrend, 0, count)
	for i := 0; i < count; i++ {
		idx := f.rng.Intn(len(trendingTopics))
		t := trendingTopics[idx]
		trends = append(trends, SocialTrend{
			Topic:      t.topic,
			Engagement: 20 + f.rng.Float64()*80,
			Sentiment:  (f.rng.Float64() * 2) - 1,
			Region:     regionID,
		})
	}
	return trends
}

func (f *FeedProvider) GetEconomicIndicators(regionID string) []EconomicIndicator {
	indicators := make([]EconomicIndicator, 0, len(economicIndicators))
	for _, name := range economicIndicators {
		indicators = append(indicators, EconomicIndicator{
			Name:   name,
			Value:  f.rng.Float64() * 100,
			Change: (f.rng.Float64() * 10) - 5,
			Region: regionID,
		})
	}
	return indicators
}

// GenerateEventContext produces a summary of current feed data that the event
// generator can use to pick contextually relevant events.
func (f *FeedProvider) GenerateEventContext() map[string]float64 {
	news := f.GetNewsItems(5)
	avgSentiment := 0.0
	categoryCounts := map[string]int{}
	for _, n := range news {
		avgSentiment += n.Sentiment
		categoryCounts[n.Category]++
	}
	avgSentiment /= float64(len(news))

	context := map[string]float64{
		"news_sentiment": avgSentiment,
	}
	for cat, count := range categoryCounts {
		context[fmt.Sprintf("news_%s_count", cat)] = float64(count)
	}
	return context
}
