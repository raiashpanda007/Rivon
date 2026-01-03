package jobs

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
)

func RunStartUpJobs(ctx context.Context, db *pgxpool.Pool, cfg *config.Config) error {
	footCfgStratupJobs := NewFootBallConfigMetaSetup(db, cfg)
	err := footCfgStratupJobs.SetupCountries(ctx)
	if err != nil {
		return err
	}
	err = footCfgStratupJobs.SetupCompetitions(ctx)
	if err != nil {
		return err
	}
	return nil
}
