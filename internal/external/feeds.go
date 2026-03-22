package external

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rhaqim/worldgame/internal/models"
)

// RSSFeed represents a parsed RSS 2.0 feed.
type RSSFeed struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []RSSItem `xml:"item"`
	} `xml:"channel"`
}

// RSSItem represents a single RSS item.
type RSSItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Category    string `xml:"category"`
}

// AtomFeed represents a parsed Atom feed.
type AtomFeed struct {
	Title   string     `xml:"title"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomEntry represents a single Atom entry.
type AtomEntry struct {
	Title   string    `xml:"title"`
	Summary string    `xml:"summary"`
	Content string    `xml:"content"`
	Links   []AtomLink `xml:"link"`
	Updated string    `xml:"updated"`
	Published string  `xml:"published"`
}

// AtomLink represents an Atom link element.
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// FeedSource defines an RSS/Atom feed URL with its tag classification.
type FeedSource struct {
	Name     string     // human-readable name
	URL      string     // feed URL
	Tag      models.Tag // primary tag
	RegionID string     // region ID if region-specific, empty for global
}

// DefaultFeedSources returns the built-in list of public RSS feeds
// organized by tag. These are all freely available without API keys.
func DefaultFeedSources() []FeedSource {
	return []FeedSource{
		// Economics
		{Name: "reuters-business", URL: "https://feeds.reuters.com/reuters/businessNews", Tag: models.TagEconomics, RegionID: ""},
		{Name: "bbc-business", URL: "https://feeds.bbci.co.uk/news/business/rss.xml", Tag: models.TagEconomics, RegionID: ""},
		{Name: "economist-finance", URL: "https://www.economist.com/finance-and-economics/rss.xml", Tag: models.TagEconomics, RegionID: ""},

		// Politics
		{Name: "reuters-politics", URL: "https://feeds.reuters.com/Reuters/PoliticsNews", Tag: models.TagPolitics, RegionID: ""},
		{Name: "bbc-politics", URL: "https://feeds.bbci.co.uk/news/politics/rss.xml", Tag: models.TagPolitics, RegionID: "uk"},
		{Name: "aljazeera", URL: "https://www.aljazeera.com/xml/rss/all.xml", Tag: models.TagPolitics, RegionID: ""},

		// Security
		{Name: "reuters-world", URL: "https://feeds.reuters.com/Reuters/worldNews", Tag: models.TagSecurity, RegionID: ""},
		{Name: "bbc-world", URL: "https://feeds.bbci.co.uk/news/world/rss.xml", Tag: models.TagSecurity, RegionID: ""},
		{Name: "threatpost", URL: "https://threatpost.com/feed/", Tag: models.TagSecurity, RegionID: ""},

		// Education
		{Name: "bbc-education", URL: "https://feeds.bbci.co.uk/news/education/rss.xml", Tag: models.TagEducation, RegionID: "uk"},
		{Name: "education-week", URL: "https://www.edweek.org/feed", Tag: models.TagEducation, RegionID: "usa"},
		{Name: "unesco", URL: "https://www.unesco.org/en/rss.xml", Tag: models.TagEducation, RegionID: ""},

		// Law
		{Name: "reuters-legal", URL: "https://feeds.reuters.com/reuters/USlegalnews", Tag: models.TagLaw, RegionID: "usa"},
		{Name: "bbc-law", URL: "https://feeds.bbci.co.uk/news/uk/rss.xml", Tag: models.TagLaw, RegionID: "uk"},
		{Name: "jurist", URL: "https://www.jurist.org/news/feed/", Tag: models.TagLaw, RegionID: ""},
	}
}

// FeedFetcher fetches and parses RSS/Atom feeds from external sources.
type FeedFetcher struct {
	client  *http.Client
	sources []FeedSource
}

// NewFeedFetcher creates a new FeedFetcher with the given sources.
func NewFeedFetcher(sources []FeedSource) *FeedFetcher {
	return &FeedFetcher{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		sources: sources,
	}
}

// FetchAll fetches all configured feeds and returns parsed feed items.
func (f *FeedFetcher) FetchAll(ctx context.Context) []models.FeedItem {
	var allItems []models.FeedItem

	for _, src := range f.sources {
		items, err := f.fetchFeed(ctx, src)
		if err != nil {
			log.Printf("[FeedFetcher] Error fetching %s (%s): %v", src.Name, src.URL, err)
			continue
		}
		allItems = append(allItems, items...)
		log.Printf("[FeedFetcher] Fetched %d items from %s", len(items), src.Name)
	}

	return allItems
}

// FetchByTag fetches feeds for a specific tag only.
func (f *FeedFetcher) FetchByTag(ctx context.Context, tag models.Tag) []models.FeedItem {
	var items []models.FeedItem

	for _, src := range f.sources {
		if src.Tag != tag {
			continue
		}
		fetched, err := f.fetchFeed(ctx, src)
		if err != nil {
			log.Printf("[FeedFetcher] Error fetching %s: %v", src.Name, err)
			continue
		}
		items = append(items, fetched...)
	}

	return items
}

// FetchByTagAndRegion fetches feeds for a specific tag, preferring region-specific feeds.
func (f *FeedFetcher) FetchByTagAndRegion(ctx context.Context, tag models.Tag, regionID string) []models.FeedItem {
	var items []models.FeedItem

	for _, src := range f.sources {
		if src.Tag != tag {
			continue
		}
		// Include global feeds and region-specific feeds for this region.
		if src.RegionID != "" && src.RegionID != regionID {
			continue
		}
		fetched, err := f.fetchFeed(ctx, src)
		if err != nil {
			log.Printf("[FeedFetcher] Error fetching %s: %v", src.Name, err)
			continue
		}
		// Tag region-specific items.
		for i := range fetched {
			if src.RegionID != "" {
				fetched[i].RegionID = src.RegionID
			}
		}
		items = append(items, fetched...)
	}

	return items
}

func (f *FeedFetcher) fetchFeed(ctx context.Context, src FeedSource) ([]models.FeedItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "WorldGame/1.0 (Feed Aggregator)")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5MB limit
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// Try RSS first, then Atom.
	items, err := parseRSS(body, src)
	if err != nil || len(items) == 0 {
		items, err = parseAtom(body, src)
		if err != nil {
			return nil, fmt.Errorf("parse feed: %w", err)
		}
	}

	return items, nil
}

func parseRSS(data []byte, src FeedSource) ([]models.FeedItem, error) {
	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	now := time.Now()
	var items []models.FeedItem

	for _, entry := range feed.Channel.Items {
		if entry.Title == "" {
			continue
		}

		pubTime := parseTime(entry.PubDate)
		if pubTime.IsZero() {
			pubTime = now
		}

		desc := stripHTML(entry.Description)
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}

		items = append(items, models.FeedItem{
			Tag:         src.Tag,
			RegionID:    src.RegionID,
			Title:       cleanText(entry.Title),
			Description: desc,
			URL:         entry.Link,
			Source:      "rss",
			FeedName:    src.Name,
			PublishedAt: pubTime,
			FetchedAt:   now,
		})
	}

	return items, nil
}

func parseAtom(data []byte, src FeedSource) ([]models.FeedItem, error) {
	var feed AtomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	now := time.Now()
	var items []models.FeedItem

	for _, entry := range feed.Entries {
		if entry.Title == "" {
			continue
		}

		pubTime := parseTime(entry.Published)
		if pubTime.IsZero() {
			pubTime = parseTime(entry.Updated)
		}
		if pubTime.IsZero() {
			pubTime = now
		}

		desc := entry.Summary
		if desc == "" {
			desc = entry.Content
		}
		desc = stripHTML(desc)
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}

		link := ""
		for _, l := range entry.Links {
			if l.Rel == "" || l.Rel == "alternate" {
				link = l.Href
				break
			}
		}

		items = append(items, models.FeedItem{
			Tag:         src.Tag,
			RegionID:    src.RegionID,
			Title:       cleanText(entry.Title),
			Description: desc,
			URL:         link,
			Source:      "rss",
			FeedName:    src.Name,
			PublishedAt: pubTime,
			FetchedAt:   now,
		})
	}

	return items, nil
}

// parseTime tries several common date formats used in RSS/Atom feeds.
func parseTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		time.RFC3339Nano,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	// Collapse whitespace.
	text := result.String()
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

// cleanText normalizes whitespace in text.
func cleanText(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// tagKeywords maps tags to keywords for re-classification of feed items.
var tagKeywords = map[models.Tag][]string{
	models.TagEconomics: {"economy", "economic", "gdp", "inflation", "trade", "market", "stock", "finance", "bank", "fiscal", "unemployment", "tariff", "export", "import", "investment", "currency", "debt", "growth", "recession", "startup"},
	models.TagPolitics:  {"election", "government", "parliament", "congress", "president", "minister", "vote", "party", "legislation", "diplomacy", "democracy", "opposition", "coalition", "reform", "policy", "campaign", "political", "senate", "governor", "mayor"},
	models.TagSecurity:  {"security", "military", "defense", "terrorism", "cyber", "attack", "threat", "intelligence", "surveillance", "crime", "weapon", "conflict", "war", "border", "police", "hacking", "ransomware", "espionage", "nato", "missile"},
	models.TagEducation: {"education", "school", "university", "student", "teacher", "curriculum", "literacy", "scholarship", "research", "academic", "learning", "college", "degree", "stem", "training", "campus", "classroom", "tuition", "graduate", "professor"},
	models.TagLaw:       {"law", "legal", "court", "justice", "judge", "legislation", "regulation", "rights", "constitutional", "criminal", "civil", "attorney", "prosecutor", "verdict", "statute", "compliance", "enforcement", "judicial", "supreme", "amendment"},
}

// ClassifyTag attempts to determine the best tag for a feed item based on
// its title and description content. Returns the original tag if no better match.
func ClassifyTag(title, description string, originalTag models.Tag) models.Tag {
	text := strings.ToLower(title + " " + description)

	scores := make(map[models.Tag]int)
	for tag, keywords := range tagKeywords {
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				scores[tag]++
			}
		}
	}

	// If the original tag scores well enough, keep it.
	if scores[originalTag] >= 2 {
		return originalTag
	}

	// Find the highest scoring tag.
	bestTag := originalTag
	bestScore := 0
	for tag, score := range scores {
		if score > bestScore {
			bestScore = score
			bestTag = tag
		}
	}

	if bestScore >= 2 {
		return bestTag
	}
	return originalTag
}

// regionKeywords maps region IDs to country/region keywords for matching feed items to regions.
var regionKeywords = map[string][]string{
	"usa":          {"united states", "u.s.", "us ", "america", "washington", "american", "biden", "trump", "congress"},
	"china":        {"china", "chinese", "beijing", "shanghai"},
	"germany":      {"germany", "german", "berlin", "bundesbank"},
	"japan":        {"japan", "japanese", "tokyo"},
	"brazil":       {"brazil", "brazilian", "brasilia", "são paulo"},
	"india":        {"india", "indian", "delhi", "mumbai", "modi"},
	"uk":           {"britain", "british", "uk ", "london", "england", "scotland", "wales"},
	"france":       {"france", "french", "paris", "macron"},
	"south_korea":  {"south korea", "korean", "seoul"},
	"australia":    {"australia", "australian", "sydney", "canberra"},
	"russia":       {"russia", "russian", "moscow", "kremlin", "putin"},
	"nigeria":      {"nigeria", "nigerian", "lagos", "abuja"},
	"south_africa": {"south africa", "johannesburg", "cape town"},
	"mexico":       {"mexico", "mexican", "mexico city"},
	"egypt":        {"egypt", "egyptian", "cairo"},
	"indonesia":    {"indonesia", "indonesian", "jakarta"},
	"saudi_arabia": {"saudi", "riyadh", "saudi arabia"},
	"canada":       {"canada", "canadian", "ottawa", "toronto"},
}

// DetectRegion tries to match a feed item to a region based on keywords in title and description.
// Returns the matching region ID, or empty string if no match.
func DetectRegion(title, description string) string {
	text := strings.ToLower(title + " " + description)

	bestRegion := ""
	bestScore := 0

	for regionID, keywords := range regionKeywords {
		score := 0
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestRegion = regionID
		}
	}

	if bestScore >= 1 {
		return bestRegion
	}
	return ""
}
