package jobs

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
)

type FootballConfigMetaDataSetup interface {
	SetupCountries(ctx context.Context) error    // --> not a cron job
	SetupCompetitions(ctx context.Context) error // --> not  a cron job
}
type footballJobUtils struct {
	FootballRepo       FootBallMetaRepo
	FootballStaticData *config.FootballStaticCompetitions
}

func NewFootBallConfigMetaSetup(db *pgxpool.Pool, cfg *config.Config) FootballConfigMetaDataSetup {
	footballrepo := NewFootBallMetaRepo(db)
	return &footballJobUtils{FootballRepo: footballrepo, FootballStaticData: &cfg.FootBallStaticData}
}

func (r *footballJobUtils) SetupCountries(ctx context.Context) error {
	competitionsStaticJsonData := r.FootballStaticData.Competitions
	for _, val := range competitionsStaticJsonData {
		err := r.FootballRepo.SaveCountries(ctx, val.Country.Name, val.Country.Emblem, val.Country.Code, val.Country.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *footballJobUtils) SetupCompetitions(ctx context.Context) error {
	competitionStaticJsonData := r.FootballStaticData.Competitions
	for _, val := range competitionStaticJsonData {
		err := r.FootballRepo.SaveCompetitions(ctx, val.Name, val.Code, val.Emblem, val.ID, val.Country.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
