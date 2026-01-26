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
	ID              uuid.UUID
	TeamID          uuid.UUID
	LeagueID        uuid.UUID
	SeasonID        uuid.UUID
	PlayedGames     int
	Won             int
	Lost            int
	Draw            int
	Points          int
	GoalsFor        int
	GoalsAgainst    int
	GoalsDifference int
	Position        int
	TeamName        string
	TeamShortName   string
	TeamCode        string
	TeamTLA         string
	TeamEmblem      string
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
