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
