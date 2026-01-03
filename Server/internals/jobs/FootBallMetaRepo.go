package jobs

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
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
