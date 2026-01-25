package jobs

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/types"
	"log"
	"time"
)

type FootBallMetaRepo interface {
	SaveCountries(ctx context.Context, name, emblem, code string, football_org_id int) error
	SaveCompetitions(ctx context.Context, name, code, emblem string, football_org_id, footBallOrg_countryID int) error
	GetCompetitions(ctx context.Context) ([]types.LeagueStruct, error)
	SaveSeasons(ctx context.Context, season string, footballOrgID int, startDate, endDate time.Time, matchDay int, winnerTeamId *uuid.UUID, leagueId uuid.UUID) (uuid.UUID, error)
	SaveTeamWithLeague(ctx context.Context, name, shortName, code, tla, emblem string, footballOrgId int, leagueId uuid.UUID, seasonId uuid.UUID) (*uuid.UUID, error)
	SaveStandings(ctx context.Context, teamId uuid.UUID, leagueId uuid.UUID, seasonId uuid.UUID, playedGames int, won int, draw int, lost int, goalsFor int, goalsAgainst int, position int) error
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
		return errors.New(errMsg)
	}
	if ct.RowsAffected() == 0 {
		errMsg := fmt.Sprintf(
			"Country not found for football_org_countryID=%d. League=%s not inserted",
			footBallOrg_countryID,
			name,
		)
		return errors.New(errMsg)
	}
	log.Printf("Saved league :: %s :: %s", name, code)
	return nil
}

func (r *footballMetaRepoServices) SaveSeasons(ctx context.Context, season string, footballOrgID int, startDate, endDate time.Time, matchDay int, winnerTeamId *uuid.UUID, leagueId uuid.UUID) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
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
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO league_seasons (league_id, season_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`, leagueId, seasonID)
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return seasonID, nil
}

func (r *footballMetaRepoServices) SaveTeamWithLeague(ctx context.Context, name, shortName, code, tla, emblem string, footballOrgId int, leagueId uuid.UUID, seasonId uuid.UUID) (*uuid.UUID, error) {
	var teamId uuid.UUID
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
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
		return nil, err
	}

	return &teamId, tx.Commit(ctx)
}

func (r *footballMetaRepoServices) GetCompetitions(ctx context.Context) ([]types.LeagueStruct, error) {
	var leagues []types.LeagueStruct
	query := `
	SELECT id, name, code, emblem, football_org_id, country_id, created_at, updated_at
	FROM leagues;
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return leagues, err
	}
	defer rows.Close()

	for rows.Next() {
		var league types.LeagueStruct
		err := rows.Scan(&league.Id, &league.Name, &league.Code, &league.Emblem, &league.FootballOrgId, &league.CountryId, &league.CreatedAt, &league.UpdatedAt)
		if err != nil {
			return leagues, err
		}
		leagues = append(leagues, league)
	}

	if err := rows.Err(); err != nil {
		return leagues, err
	}
	return leagues, nil
}

func (r *footballMetaRepoServices) SaveStandings(ctx context.Context, teamId uuid.UUID, leagueId uuid.UUID, seasonId uuid.UUID, playedGames int, won int, draw int, lost int, goalsFor int, goalsAgainst int, position int) error {
	if won+draw+lost != playedGames {
		return fmt.Errorf("invalid standings: won + draw + lost must equal played games")
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

	return err
}

func (r *footballMetaRepoServices) GetStandings() error {

	return nil
}
