package jobs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/types"
)

type FootBallMetaRepo interface {
	SaveCountries(ctx context.Context, name, emblem, code string, football_org_id int) error
	SaveCompetitions(ctx context.Context, name, code, emblem string, football_org_id, footBallOrg_countryID int) error
	GetCompetitions(ctx context.Context) ([]types.LeagueStruct, error)
	SaveSeasons(ctx context.Context, season string, footballOrgID int, startDate, endDate time.Time, matchDay int, winnerTeamId *uuid.UUID, leagueId uuid.UUID) (uuid.UUID, error)
	SaveTeamWithLeague(ctx context.Context, name, shortName, code, tla, emblem string, footballOrgId int, leagueId uuid.UUID, seasonId uuid.UUID) (*uuid.UUID, error)
	SaveStandings(ctx context.Context, teamId uuid.UUID, leagueId uuid.UUID, seasonId uuid.UUID, playedGames int, won int, draw int, lost int, goalsFor int, goalsAgainst int, position int) error
	GetCompetitionTeamStandings(ctx context.Context, leagueId, seasonId *uuid.UUID) ([]types.StandingsQueryResponse, error)
	GetCompetitionMetaData(ctx context.Context, leagueId uuid.UUID) (types.GetCompetitionMetaData, error)
	GetAllCompetitionMetaData(ctx context.Context) ([]types.GetCompetitionMetaData, error)
	GetAllSeasons(ctx context.Context) ([]types.GetSeason, error)
	GetAllLeagueSeasons(ctx context.Context) ([]types.GetLeagueSeason, error)
}

type footballMetaRepoServices struct {
	db *pgxpool.Pool
}

func NewFootBallMetaRepo(db *pgxpool.Pool) FootBallMetaRepo {
	return &footballMetaRepoServices{
		db: db,
	}
}
func (r *footballMetaRepoServices) SaveCountries(ctx context.Context, name, emblem, code string, football_org_id int) error {
	query := `
		INSERT INTO countries (id , name , code , emblem, football_org_id) 
		VALUES ($1 , $2, $3, $4, $5)
		ON CONFLICT (football_org_id) DO NOTHING;
	`
	ct, err := r.db.Exec(ctx, query, uuid.New(), name, code, emblem, football_org_id)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to save country in DB :: %s :: code :: %s :: Error :: %s", name, code, err.Error())
		slog.Error("Unable to save country in DB", "error", err)
		return errors.New(errMsg)
	}
	if ct.RowsAffected() == 0 {
		log.Printf("Country already exists :: %s :: %s", name, code)
		return nil
	}
	log.Printf("Saved country :: %s :: %s", name, code)
	return nil
}

func (r *footballMetaRepoServices) SaveCompetitions(ctx context.Context, name, code, emblem string, footballOrgID, footBallOrg_countryID int) error {

	query := `
    INSERT INTO leagues (
        id, name, code, emblem, football_org_id, country_id
    )
    SELECT
        $1, $2, $3, $4, $5, c.id
    FROM countries c
    WHERE c.football_org_id = $6
		ON CONFLICT (football_org_id) DO NOTHING;
    `

	ct, err := r.db.Exec(
		ctx,
		query,
		uuid.New(),
		name,
		code,
		emblem,
		footballOrgID,
		footBallOrg_countryID,
	)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to save country in DB :: %s :: code :: %s :: Error :: %s", name, code, err.Error())
		slog.Error("Unable to save league in DB", "error", err)
		return errors.New(errMsg)
	}
	if ct.RowsAffected() == 0 {
		errMsg := fmt.Sprintf(
			"Country not found for football_org_countryID=%d. League=%s not inserted",
			footBallOrg_countryID,
			name,
		)
		slog.Error("Country not found for football_org_countryID", "error", errors.New(errMsg))
		return errors.New(errMsg)
	}
	log.Printf("Saved league :: %s :: %s", name, code)
	return nil
}

func (r *footballMetaRepoServices) SaveSeasons(ctx context.Context, season string, footballOrgID int, startDate, endDate time.Time, matchDay int, winnerTeamId *uuid.UUID, leagueId uuid.UUID) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		slog.Error("Error starting transaction", "error", err)
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var seasonID uuid.UUID

	err = tx.QueryRow(ctx, `
		INSERT INTO seasons (
			id,
			football_org_id,
			season,
			period,
			winner_team_id,
			match_day
		)
		VALUES (
			$1,
			$2,
			$3,
			daterange($4, $5, '[]'),
			$6,
			$7
		)
		ON CONFLICT (football_org_id)
		DO UPDATE SET
			season = EXCLUDED.season,
			period = EXCLUDED.period,
			winner_team_id = EXCLUDED.winner_team_id,
			match_day = EXCLUDED.match_day,
			updated_at = NOW()
		RETURNING id;
	`,
		uuid.New(),
		footballOrgID,
		season,
		startDate,
		endDate,
		winnerTeamId,
		matchDay,
	).Scan(&seasonID)

	if err != nil {
		slog.Error("Error saving season", "error", err)
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO league_seasons (league_id, season_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`, leagueId, seasonID)
	if err != nil {
		slog.Error("Error saving league season", "error", err)
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return uuid.Nil, err
	}
	return seasonID, nil
}

func (r *footballMetaRepoServices) SaveTeamWithLeague(ctx context.Context, name, shortName, code, tla, emblem string, footballOrgId int, leagueId uuid.UUID, seasonId uuid.UUID) (*uuid.UUID, error) {
	var teamId uuid.UUID
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		slog.Error("Error starting transaction", "error", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO teams (
			id,
			name,
			short_name,
			code,
			tla,
			emblem,
			football_org_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (football_org_id)
		DO UPDATE SET
			name = EXCLUDED.name,
			short_name = EXCLUDED.short_name,
			code = EXCLUDED.code,
			tla = EXCLUDED.tla,
			emblem = EXCLUDED.emblem,
			updated_at = NOW()
		RETURNING id;
	`,
		uuid.New(),
		name,
		shortName,
		code,
		tla,
		emblem,
		footballOrgId,
	).Scan(&teamId)

	if err != nil {
		slog.Error("Error saving team", "error", err)
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO teams_leagues (
			team_id,
			league_id,
			season_id
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, league_id, season_id)
		DO NOTHING;
	`,
		teamId,
		leagueId,
		seasonId,
	)
	if err != nil {
		slog.Error("Error saving team league", "error", err)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return nil, err
	}

	return &teamId, nil
}

func (r *footballMetaRepoServices) GetCompetitions(ctx context.Context) ([]types.LeagueStruct, error) {
	var leagues []types.LeagueStruct
	query := `
	SELECT id, name, code, emblem, football_org_id, country_id, created_at, updated_at
	FROM leagues;
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		slog.Error("Error querying leagues", "error", err)
		return leagues, err
	}
	defer rows.Close()

	for rows.Next() {
		var league types.LeagueStruct
		err := rows.Scan(&league.Id, &league.Name, &league.Code, &league.Emblem, &league.FootballOrgId, &league.CountryId, &league.CreatedAt, &league.UpdatedAt)
		if err != nil {
			slog.Error("Error scanning league", "error", err)
			return leagues, err
		}
		leagues = append(leagues, league)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error iterating leagues", "error", err)
		return leagues, err
	}
	return leagues, nil
}

func (r *footballMetaRepoServices) SaveStandings(ctx context.Context, teamId uuid.UUID, leagueId uuid.UUID, seasonId uuid.UUID, playedGames int, won int, draw int, lost int, goalsFor int, goalsAgainst int, position int) error {
	if won+draw+lost != playedGames {
		err := fmt.Errorf("invalid standings: won + draw + lost must equal played games")
		slog.Error("Invalid standings", "error", err)
		return err
	}

	points := (won * 3) + draw
	goalDifference := goalsFor - goalsAgainst

	query := `
	INSERT INTO standings (
		id,
		team_id,
		league_id,
		season_id,
		played_games,
		won,
		draw,
		lost,
		points,
		goals_for,
		goals_against,
		goal_difference,
		position
	)
	VALUES (
		$1, $2, $3, $4,
		$5, $6, $7, $8,
		$9, $10, $11, $12, $13
	)
	ON CONFLICT (team_id, league_id, season_id)
	DO UPDATE SET
		played_games = EXCLUDED.played_games,
		won = EXCLUDED.won,
		draw = EXCLUDED.draw,
		lost = EXCLUDED.lost,
		points = EXCLUDED.points,
		goals_for = EXCLUDED.goals_for,
		goals_against = EXCLUDED.goals_against,
		goal_difference = EXCLUDED.goal_difference,
		position = EXCLUDED.position,
		updated_at = NOW();
	`

	_, err := r.db.Exec(
		ctx,
		query,
		uuid.New(),
		teamId,
		leagueId,
		seasonId,
		playedGames,
		won,
		draw,
		lost,
		points,
		goalsFor,
		goalsAgainst,
		goalDifference,
		position,
	)
	if err != nil {
		slog.Error("Error saving standings", "error", err)
	}

	return err
}

func (r *footballMetaRepoServices) GetCompetitionTeamStandings(ctx context.Context, leagueId, seasonId *uuid.UUID) ([]types.StandingsQueryResponse, error) {
	var standings []types.StandingsQueryResponse

	query := `
	SELECT
	  s.id,
	  s.team_id,
	  s.league_id,
	  s.season_id,
	  s.played_games,
	  s.won,
	  s.lost,
	  s.draw,
	  s.points,
	  s.goals_for,
	  s.goals_against,
	  s.goal_difference,
	  s.position,

	  t.name,
	  t.short_name,
	  t.code,
	  t.tla,
	  t.emblem

	FROM standings s
	JOIN teams t ON t.id = s.team_id
	WHERE
	  ($1::uuid IS NULL OR s.league_id = $1)
	  AND
	  ($2::uuid IS NULL OR s.season_id = $2)
	ORDER BY s.position ASC;
	`

	rows, err := r.db.Query(ctx, query, leagueId, seasonId)
	if err != nil {
		slog.Error("Error querying standings", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var standing types.StandingsQueryResponse

		err := rows.Scan(
			&standing.ID,
			&standing.TeamID,
			&standing.LeagueID,
			&standing.SeasonID,
			&standing.PlayedGames,
			&standing.Won,
			&standing.Lost,
			&standing.Draw,
			&standing.Points,
			&standing.GoalsFor,
			&standing.GoalsAgainst,
			&standing.GoalsDifference,
			&standing.Position,

			&standing.TeamName,
			&standing.TeamShortName,
			&standing.TeamCode,
			&standing.TeamTLA,
			&standing.TeamEmblem,
		)
		if err != nil {
			slog.Error("Error scanning standing", "error", err)
			return nil, err
		}

		standings = append(standings, standing)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error iterating standings", "error", err)
		return nil, err
	}

	return standings, nil
}

func (r *footballMetaRepoServices) GetCompetitionMetaData(ctx context.Context, leagueId uuid.UUID) (types.GetCompetitionMetaData, error) {
	query := `
SELECT
    l.id              AS league_id,
    l.name            AS league_name,
    l.code            AS league_code,
    l.emblem          AS league_emblem,
    l.football_org_id AS league_football_org_id,

    c.id              AS country_id,
    c.name            AS country_name,
    c.code            AS country_code,
    c.emblem          AS country_emblem,
    c.football_org_id AS country_football_org_id
FROM leagues l
INNER JOIN countries c
    ON l.country_id = c.id
WHERE l.id = $1;
	`
	var competitionDetails types.GetCompetitionMetaData
	err := r.db.QueryRow(ctx, query, leagueId).Scan(&competitionDetails.ID, &competitionDetails.Name, &competitionDetails.Code, &competitionDetails.Emblem, &competitionDetails.FootballOrgId, &competitionDetails.CountryId, &competitionDetails.CountryName, &competitionDetails.CountryCode, &competitionDetails.CountryEmblem, &competitionDetails.CountryFootBallOrgCode)

	if err != nil {
		slog.Error("Error getting competition meta data", "error", err)
	}
	return competitionDetails, err
}

func (r *footballMetaRepoServices) GetAllCompetitionMetaData(ctx context.Context) ([]types.GetCompetitionMetaData, error) {
	query := `
SELECT
    l.id              AS league_id,
    l.name            AS league_name,
    l.code            AS league_code,
    l.emblem          AS league_emblem,
    l.football_org_id AS league_football_org_id,

    c.id              AS country_id,
    c.name            AS country_name,
    c.code            AS country_code,
    c.emblem          AS country_emblem,
    c.football_org_id AS country_football_org_id
FROM leagues l
INNER JOIN countries c
    ON l.country_id = c.id
`
	var allCompetitionDetails []types.GetCompetitionMetaData
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		slog.Error("Error querying competition meta data", "error", err)
		return allCompetitionDetails, err
	}

	defer rows.Close()
	for rows.Next() {
		var competitionDetails types.GetCompetitionMetaData
		err := rows.Scan(&competitionDetails.ID, &competitionDetails.Name, &competitionDetails.Code, &competitionDetails.Emblem, &competitionDetails.FootballOrgId, &competitionDetails.CountryId, &competitionDetails.CountryName, &competitionDetails.CountryCode, &competitionDetails.CountryEmblem, &competitionDetails.CountryFootBallOrgCode)
		if err != nil {
			slog.Error("Error scanning competition meta data", "error", err)
			return allCompetitionDetails, err
		}

		allCompetitionDetails = append(allCompetitionDetails, competitionDetails)
	}
	if err := rows.Err(); err != nil {
		slog.Error("Error iterating competition meta data", "error", err)
		return allCompetitionDetails, err
	}
	return allCompetitionDetails, nil

}

func (r *footballMetaRepoServices) GetAllSeasons(ctx context.Context) ([]types.GetSeason, error) {
	var allSeasons []types.GetSeason
	query := `
		SELECT id , season, lower(period), upper(period), match_day, winner_team_id, created_at, updated_at
		FROM seasons;
 	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		slog.Error("Error in getting all seasons details", "error", err)
		return allSeasons, err
	}
	defer rows.Close()

	for rows.Next() {
		var season types.GetSeason
		err := rows.Scan(&season.ID, &season.Season, &season.Period.Start, &season.Period.End, &season.MatchDay, &season.WinnerTeamID, &season.CreatedAt, &season.UpdatedAt)

		if err != nil {
			slog.Error("Error scanning competition meta data", "error", err)
			return allSeasons, err
		}
		allSeasons = append(allSeasons, season)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error iterating competition meta data", "error", err)
		return allSeasons, err
	}

	return allSeasons, nil

}

func (r *footballMetaRepoServices) GetAllLeagueSeasons(ctx context.Context) ([]types.GetLeagueSeason, error) {
	var allLeagueSeason []types.GetLeagueSeason
	query := `
	SELECT league_id, season_id 
	FROM league_seasons;
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return allLeagueSeason, err
	}
	defer rows.Close()

	for rows.Next() {
		var leagueSeason types.GetLeagueSeason
		err := rows.Scan(&leagueSeason.LeagueID, &leagueSeason.SeasonID)
		if err != nil {
			slog.Error("Error scanning league season ", "error ", err)
			return allLeagueSeason, err
		}
		allLeagueSeason = append(allLeagueSeason, leagueSeason)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error in iterating league seasons", "error", err)
		return allLeagueSeason, err
	}

	return allLeagueSeason, nil

}
