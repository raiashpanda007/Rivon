package types

import (
	"time"

	"github.com/google/uuid"
)

type RegisterType struct {
	Name     string `json:"name" validator:"required"`
	Password string `json:"password" validator:"required"`
	Email    string `json:"email" validator:"required"`
}

type LoginType struct {
	Email    string `json:"email" validator:"required"`
	Password string `json:"password" validator:"required"`
}

type RefreshTokenType struct {
	Id string `json:"id" validator:"required"`
}
type VerifyOTPCredentials struct {
	OTP string `json:"otp" validator:"required"`
}

type TransactionType string

const (
	DEBIT  TransactionType = "debit"
	CREDIT TransactionType = "credit"
)

type LeagueStruct struct {
	Id            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Code          string    `json:"code"`
	Emblem        string    `json:"emblem"`
	FootballOrgId int       `json:"football_org_id"`
	CountryId     uuid.UUID `json:"country"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StandingsResponse struct {
	Filters     Filters     `json:"filters"`
	Area        Area        `json:"area"`
	Competition Competition `json:"competition"`
	Season      Season      `json:"season"`
	Standings   []Standing  `json:"standings"`
}

type Filters struct {
	Season string `json:"season"`
}

type Area struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Flag string `json:"flag"`
}

type Competition struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Type   string `json:"type"`
	Emblem string `json:"emblem"`
}

type Season struct {
	ID              int    `json:"id"`
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	CurrentMatchday int    `json:"currentMatchday"`
	Winner          any    `json:"winner"` // null → future-proof
}

type Standing struct {
	Stage string       `json:"stage"`
	Type  string       `json:"type"`
	Group *string      `json:"group"` // null
	Table []TableEntry `json:"table"`
}

type TableEntry struct {
	Position       int     `json:"position"`
	Team           Team    `json:"team"`
	PlayedGames    int     `json:"playedGames"`
	Form           *string `json:"form"` // null
	Won            int     `json:"won"`
	Draw           int     `json:"draw"`
	Lost           int     `json:"lost"`
	Points         int     `json:"points"`
	GoalsFor       int     `json:"goalsFor"`
	GoalsAgainst   int     `json:"goalsAgainst"`
	GoalDifference int     `json:"goalDifference"`
}

type Team struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	TLA       string `json:"tla"`
	Crest     string `json:"crest"`
}

type StandingsQueryResponse struct {
	ID              uuid.UUID `json:"id"`
	TeamID          uuid.UUID `json:"teamId"`
	LeagueID        uuid.UUID `json:"leagueId"`
	SeasonID        uuid.UUID `json:"seasonId"`
	PlayedGames     int       `json:"playedGames"`
	Won             int       `json:"won"`
	Lost            int       `json:"lost"`
	Draw            int       `json:"draw"`
	Points          int       `json:"points"`
	GoalsFor        int       `json:"goalsFor"`
	GoalsAgainst    int       `json:"goalsAgainst"`
	GoalsDifference int       `json:"goalDifference"`
	Position        int       `json:"position"`
	TeamName        string    `json:"teamName"`
	TeamShortName   string    `json:"teamShortName"`
	TeamCode        string    `json:"teamCode"`
	TeamTLA         string    `json:"teamTLA"`
	TeamEmblem      string    `json:"teamEmblem"`
}

type GetCompetitionMetaData struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	Code                   string    `json:"code"`
	Emblem                 string    `json:"emblem"`
	FootballOrgId          int       `json:"football_org_id"`
	CountryId              uuid.UUID `json:"countryId"`
	CountryName            string    `json:"countryName"`
	CountryFootBallOrgCode int       `json:"countryFootBallOrgCode"`
	CountryCode            string    `json:"countryCode"`
	CountryEmblem          string    `json:"countryEmblem"`
}

type DateRangeJSON struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type GetSeason struct {
	ID           uuid.UUID     `json:"id"`
	Period       DateRangeJSON `json:"period"`
	Season       string        `json:"season"`
	MatchDay     int           `json:"matchDay"`
	WinnerTeamID *uuid.UUID    `json:"winner"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type GetLeagueSeason struct {
	LeagueID uuid.UUID `json:"leagueId"`
	SeasonID uuid.UUID `json:"seasonId"`
}

type GetSeasonResponse struct {
	Seasons      []GetSeason       `json:"season"`
	LeagueSeason []GetLeagueSeason `json:"leagueSeason"`
}

type MarketTable struct {
	Id           uuid.UUID    `json:"id"`
	TeamID       uuid.UUID    `json:"teamId"`
	MarketName   string       `json:"marketName"`
	MarketCode   string       `json:"marketCode"`
	LastPrice    int64        `json:"lastPrice"`
	MarketStatus string       `json:"status"`
	Volume24H    int64        `json:"volume24h"`
	TotalVolume  int64        `json:"totalVolume"`
	OpenPrice24H int64        `json:"openPrice"`
	TeamDetails  *TeamDetails `json:"teamDetails,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

type TeamDetails struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	ShortName     string    `json:"shortName"`
	Code          string    `json:"code"`
	TLA           string    `json:"tla"`
	Emblem        string    `json:"emblem"`
	FootballOrgId int       `json:"footballOrgId"`
}

type OrderTypes string

const (
	BUY_ORDER    OrderTypes = "BUY"
	SELL_ORDER   OrderTypes = "SELL"
	CANCEL_ORDER OrderTypes = "CANCEL_ORDER"
)

type RedisStreamMessage struct {
	UserId   uuid.UUID `json:"userId"`
	MarketId uuid.UUID `json:"marketId"`
	Price    int64     `json:"price"`
	Quantity int       `json:"quantity"`
	OrderId  uuid.UUID `json:"orderId"`
}

type MarketOrder struct {
	MarketId  uuid.UUID  `json:"marketId"`
	Price     int64      `json:"price"`
	Quantity  int64      `json:"quantity"`
	OrderType OrderTypes `json:"orderType"`
	OrderId   *uuid.UUID `json:"orderId,omitempty"` // required for CANCEL_ORDER
}

type Fills struct {
	Price        int    `json:"price"`
	Quantity     int    `json:"quantity"`
	TradeId      string `json:"tradeId"`
	OtherUserId  string `json:"otherUserId"`
	OtherOrderId string `json:"otherOrderId"`
	OrderId      string `json:"orderId"`
}

type FillResult struct {
	OrderId          string  `json:"orderId"`
	ExecutedQuantity int     `json:"executedQty"`
	Fills            []Fills `json:"fills"`
	Error            string  `json:"error,omitempty"`
}

// PubSubOrderMessage mirrors the Engine's JSON payload on channel "ORDERS".
// NOTE: Engine must json.Marshal the message before calling redisClient.Publish;
// go-redis v8 uses fmt.Sprint on plain structs, not JSON. JSON tags must match
// Engine's api.pubsub.go.
type PubSubOrderMessage struct {
	OrderId          string  `json:"orderId"`
	Fills            []Fills `json:"fills"`
	ExecutedQuantity int     `json:"executedQty"`
	MessageType      string  `json:"type"`
	Error            string  `json:"error,omitempty"`
}

// MatchesResponse is the football-data.org /v4/competitions/{id}/matches response.
type MatchesResponse struct {
	Competition Competition  `json:"competition"`
	Season      Season       `json:"season"`
	Matches     []MatchEntry `json:"matches"`
}

type MatchEntry struct {
	ID       int        `json:"id"`
	Stage    string     `json:"stage"`
	Status   string     `json:"status"` // SCHEDULED, IN_PLAY, PAUSED, FINISHED, POSTPONED, CANCELLED, TIMED
	HomeTeam MatchTeam  `json:"homeTeam"`
	AwayTeam MatchTeam  `json:"awayTeam"`
	Score    MatchScore `json:"score"`
}

type MatchTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	TLA       string `json:"tla"`
	Crest     string `json:"crest"`
}

type MatchScore struct {
	Winner   *string   `json:"winner"` // HOME_TEAM, AWAY_TEAM, DRAW, nil
	Duration string    `json:"duration"`
	FullTime MatchHalf `json:"fullTime"`
	HalfTime MatchHalf `json:"halfTime"`
}

type MatchHalf struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// KnockoutMatchRow is one raw row from the knockout_matches table (with team joins).
type KnockoutMatchRow struct {
	ID                 uuid.UUID `json:"id"`
	FootballOrgMatchID int       `json:"footballOrgMatchId"`
	Stage              string    `json:"stage"`
	HomeTeamID         uuid.UUID `json:"homeTeamId"`
	AwayTeamID         uuid.UUID `json:"awayTeamId"`
	HomeTeamName       string    `json:"homeTeamName"`
	HomeTeamShortName  string    `json:"homeTeamShortName"`
	HomeTeamTLA        string    `json:"homeTeamTLA"`
	HomeTeamEmblem     string    `json:"homeTeamEmblem"`
	AwayTeamName       string    `json:"awayTeamName"`
	AwayTeamShortName  string    `json:"awayTeamShortName"`
	AwayTeamTLA        string    `json:"awayTeamTLA"`
	AwayTeamEmblem     string    `json:"awayTeamEmblem"`
	HomeScore          *int      `json:"homeScore"`
	AwayScore          *int      `json:"awayScore"`
	Status             string    `json:"status"`
}

type KnockoutTeamInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"shortName"`
	TLA       string    `json:"tla"`
	Emblem    string    `json:"emblem"`
}

type KnockoutLeg struct {
	HomeTeamID uuid.UUID `json:"homeTeamId"`
	HomeScore  *int      `json:"homeScore"`
	AwayScore  *int      `json:"awayScore"`
	Status     string    `json:"status"`
}

type KnockoutMatchup struct {
	Team1         KnockoutTeamInfo `json:"team1"`
	Team2         KnockoutTeamInfo `json:"team2"`
	Leg1          *KnockoutLeg     `json:"leg1"`
	Leg2          *KnockoutLeg     `json:"leg2"`
	Team1AggGoals *int             `json:"team1AggGoals"`
	Team2AggGoals *int             `json:"team2AggGoals"`
	WinnerTeamID  *uuid.UUID       `json:"winnerTeamId"`
}

type KnockoutStageData struct {
	Stage    string            `json:"stage"`
	Matchups []KnockoutMatchup `json:"matchups"`
}

type PlaceOrderResponse struct {
	OrderId          string  `json:"orderId"`
	ExecutedQuantity int     `json:"executedQty"`
	Fills            []Fills `json:"fills"`
	Status           string  `json:"status"`
	Message          string  `json:"message"`
}
