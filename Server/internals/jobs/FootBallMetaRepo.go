package jobs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FootBallMetaRepo interface {
	SaveCountries(ctx context.Context, name, emblem, code string, football_org_id int) error
	SaveCompetitions(ctx context.Context, name, code, emblem string, football_org_id, footBallOrg_countryID int) error
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

func (r *footballMetaRepoServices) SaveSeasons(ctx context.Context, season string, footballOrgID int, startDate, endDate time.Time, matchDay int, winnerTeamId *uuid.UUID, leagueId uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
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
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO league_seasons (league_id, season_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`, leagueId, seasonID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *footballMetaRepoServices) SaveTeamWithLeague(ctx context.Context, name, shortName, code, tla, emblem string, footballOrgId int, leagueId uuid.UUID, seasonId uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var teamId uuid.UUID

	// 1️⃣ Insert or update team
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
		return err
	}

	// 2️⃣ Insert team–league mapping
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
		return err
	}

	return tx.Commit(ctx)
}
