package models

import "time"

type SectorType string

const (
	Economics SectorType = "economics"
	Politics  SectorType = "politics"
	Security  SectorType = "security"
	Education SectorType = "education"
	RandD     SectorType = "randd"
)

var AllSectors = []SectorType{Economics, Politics, Security, Education, RandD}

type GamePhase string

const (
	PhaseLobby    GamePhase = "lobby"
	PhasePlaying  GamePhase = "playing"
	PhaseFinished GamePhase = "finished"
)

type ActionType string

const (
	ActionAllocate     ActionType = "allocate"
	ActionPolicy       ActionType = "policy"
	ActionTrade        ActionType = "trade"
	ActionRespondEvent ActionType = "respond_event"
)

type Resources struct {
	Budget    float64 `json:"budget"`
	Influence float64 `json:"influence"`
	Stability float64 `json:"stability"`
	Knowledge float64 `json:"knowledge"`
}

type SectorState struct {
	Level      float64 `json:"level"`
	Investment float64 `json:"investment"`
	Growth     float64 `json:"growth"`
}

type Player struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RegionID  string    `json:"region_id"`
	Resources Resources `json:"resources"`
	Ready     bool      `json:"ready"`
	Connected bool      `json:"connected"`
}

type PlayerState struct {
	Player      Player                      `json:"player"`
	Sectors     map[SectorType]*SectorState `json:"sectors"`
	Score       float64                     `json:"score"`
	TurnActions []Action                    `json:"turn_actions"`
}

type Region struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Country     string                 `json:"country"`
	Continent   string                 `json:"continent"`
	BaseStats   map[SectorType]float64 `json:"base_stats"`
	Description string                 `json:"description"`
}

type WorldEvent struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Source          string                 `json:"source"`
	AffectedSectors []SectorType           `json:"affected_sectors"`
	Impact          map[SectorType]float64 `json:"impact"`
	Duration        int                    `json:"duration"`
	RegionSpecific  string                 `json:"region_specific"`
}

type Game struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	MaxPlayers   int                     `json:"max_players"`
	Players      map[string]*PlayerState `json:"players"`
	Phase        GamePhase               `json:"phase"`
	CurrentTurn  int                     `json:"current_turn"`
	MaxTurns     int                     `json:"max_turns"`
	Events       []WorldEvent            `json:"events"`
	ActiveEvents []WorldEvent            `json:"active_events"`
	CreatedAt    time.Time               `json:"created_at"`
	HostID       string                  `json:"host_id"`
}

type Action struct {
	Type     ActionType             `json:"type"`
	PlayerID string                 `json:"player_id"`
	Data     map[string]interface{} `json:"data"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// CreateGameRequest is the body for POST /api/games.
// Only creates the game shell -- no player is added.
type CreateGameRequest struct {
	Name       string `json:"name"`
	MaxPlayers int    `json:"max_players"`
	MaxTurns   int    `json:"max_turns"`
}

// JoinGameRequest is the body for POST /api/games/:id/join.
type JoinGameRequest struct {
	PlayerName string `json:"player_name"`
	RegionID   string `json:"region_id"`
}

// GameSummary is returned in the list-games endpoint.
type GameSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Phase       GamePhase `json:"phase"`
	PlayerCount int       `json:"player_count"`
	MaxPlayers  int       `json:"max_players"`
	MaxTurns    int       `json:"max_turns"`
	CurrentTurn int       `json:"current_turn"`
	CreatedAt   time.Time `json:"created_at"`
}
