package models

import "time"

// Tag represents a topic category for challenges.
type Tag string

const (
	TagEconomics Tag = "economics"
	TagPolitics  Tag = "politics"
	TagSecurity  Tag = "security"
	TagEducation Tag = "education"
	TagLaw       Tag = "law"
)

// AllTags contains every valid tag.
var AllTags = []Tag{TagEconomics, TagPolitics, TagSecurity, TagEducation, TagLaw}

// IsValidTag returns true if the given tag is one of the known tags.
func IsValidTag(t Tag) bool {
	for _, valid := range AllTags {
		if t == valid {
			return true
		}
	}
	return false
}

// GamePhase represents the current phase of a game week.
type GamePhase string

const (
	PhaseActive     GamePhase = "active"
	PhaseEvaluating GamePhase = "evaluating"
	PhaseCompleted  GamePhase = "completed"
)

// Player represents a participant in a game.
type Player struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Points     float64 `json:"points"`      // available points this week (starts at 100)
	TotalScore float64 `json:"total_score"` // cumulative across weeks
	Connected  bool    `json:"connected"`
}

// Region represents a geographical area that a game is set in.
type Region struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	Continent   string `json:"continent"`
	Description string `json:"description"`
}

// Challenge represents a problem generated from feeds that players must solve.
type Challenge struct {
	ID          string    `json:"id"`
	Tag         Tag       `json:"tag"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Source      string    `json:"source"`     // "news", "social_media", "report"
	Region      string    `json:"region"`     // region name
	Severity    int       `json:"severity"`   // 1-10
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `json:"active"`
}

// Proposal represents a player's submitted solution to a challenge.
type Proposal struct {
	ID             string    `json:"id"`
	PlayerID       string    `json:"player_id"`
	PlayerName     string    `json:"player_name"`
	ChallengeID    string    `json:"challenge_id"`
	Description    string    `json:"description"`     // the player's proposed solution
	PointsInvested float64   `json:"points_invested"`
	SubmittedAt    time.Time `json:"submitted_at"`
	AIScore        float64   `json:"ai_score"`    // 0 until evaluated
	AIFeedback     string    `json:"ai_feedback"` // empty until evaluated
}

// WeekWinner holds the result of a weekly evaluation.
type WeekWinner struct {
	PlayerID   string  `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Score      float64 `json:"score"`
	Summary    string  `json:"summary"`
}

// Game is the top-level game state.
type Game struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	RegionID   string             `json:"region_id"`
	RegionName string             `json:"region_name"`
	Tags       []Tag              `json:"tags"`
	Phase      GamePhase          `json:"phase"`
	Players    map[string]*Player `json:"players"`
	Challenges []Challenge        `json:"challenges"`
	Proposals  []Proposal         `json:"proposals"`
	WeekNumber int                `json:"week_number"`
	WeekStart  time.Time          `json:"week_start"`
	WeekEnd    time.Time          `json:"week_end"`
	HostID     string             `json:"host_id"`
	CreatedAt  time.Time          `json:"created_at"`
	Winner     *WeekWinner        `json:"winner,omitempty"`
}

// GameSummary is a lightweight representation for listing games.
type GameSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RegionName  string    `json:"region_name"`
	Tags        []Tag     `json:"tags"`
	Phase       GamePhase `json:"phase"`
	PlayerCount int       `json:"player_count"`
	WeekNumber  int       `json:"week_number"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateGameRequest is the body for POST /api/games.
type CreateGameRequest struct {
	Name     string `json:"name"`
	RegionID string `json:"region_id"`
	Tags     []Tag  `json:"tags"`
}

// JoinGameRequest is the body for POST /api/games/:id/join.
type JoinGameRequest struct {
	PlayerName string `json:"player_name"`
}

// SubmitProposalRequest is the body for POST /api/games/:id/proposals.
type SubmitProposalRequest struct {
	PlayerID       string  `json:"player_id"`
	ChallengeID    string  `json:"challenge_id"`
	Description    string  `json:"description"`
	PointsInvested float64 `json:"points_invested"`
}

// ChallengeTemplate holds a reusable challenge skeleton stored in the database.
type ChallengeTemplate struct {
	ID                  int       `json:"id"`
	Tag                 Tag       `json:"tag"`
	TitleTemplate       string    `json:"title_template"`
	DescriptionTemplate string    `json:"description_template"`
	Source              string    `json:"source"`
	CreatedAt           time.Time `json:"created_at"`
}

// ChatMessage represents a persisted chat message in a game.
type ChatMessage struct {
	ID         int       `json:"id"`
	GameID     string    `json:"game_id"`
	PlayerID   string    `json:"player_id"`
	PlayerName string    `json:"player_name"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

// FeedItem represents a news article or report fetched from an external feed.
type FeedItem struct {
	ID          int       `json:"id"`
	Tag         Tag       `json:"tag"`
	RegionID    string    `json:"region_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`    // "rss", "api"
	FeedName    string    `json:"feed_name"` // e.g. "reuters", "bbc"
	PublishedAt time.Time `json:"published_at"`
	FetchedAt   time.Time `json:"fetched_at"`
	UsedInGame  bool      `json:"used_in_game"`
	Dismissed   bool      `json:"dismissed"`
}

// CuratedChallenge represents a challenge curated by the Game Master from a feed item.
type CuratedChallenge struct {
	ID           int       `json:"id"`
	FeedItemID   *int      `json:"feed_item_id,omitempty"`
	Tag          Tag       `json:"tag"`
	RegionID     string    `json:"region_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Source       string    `json:"source"`
	Severity     int       `json:"severity"`
	Active       bool      `json:"active"`
	UsedInGame   bool      `json:"used_in_game"`
	CuratorNotes string    `json:"curator_notes"`
	CreatedAt    time.Time `json:"created_at"`
}

// CurateChallengeRequest is the body for POST /api/game-master/curate.
type CurateChallengeRequest struct {
	FeedItemID   *int   `json:"feed_item_id,omitempty"`
	Tag          Tag    `json:"tag"`
	RegionID     string `json:"region_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Source       string `json:"source"`
	Severity     int    `json:"severity"`
	CuratorNotes string `json:"curator_notes"`
}

// WSMessage is the envelope for all WebSocket communication.
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
