package footballmeta

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/jobs"
	"github.com/raiashpanda007/rivon/internals/types"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type FootballMetaServices interface {
	GetCompetitions(ctx context.Context, leagueId *uuid.UUID) ([]types.GetCompetitionMetaData, utils.ErrorType, error)
	GetCompetitionTeamStandings(ctx context.Context, leagueId, seasonId *uuid.UUID) ([]types.StandingsQueryResponse, utils.ErrorType, error)
}

type footballRepo struct {
	repo jobs.FootBallMetaRepo
}

func InitFootBallMetaServices(db *pgxpool.Pool) FootballMetaServices {
	repo := jobs.NewFootBallMetaRepo(db)
	return &footballRepo{
		repo: repo,
	}
}

func (r *footballRepo) GetCompetitions(ctx context.Context, leagueId *uuid.UUID) ([]types.GetCompetitionMetaData, utils.ErrorType, error) {

	if leagueId == nil {
		results, err := r.repo.GetAllCompetitionMetaData(ctx)
		if err != nil {
			slog.Error("Error getting all competition meta data", "error", err)
			return results, utils.ErrInternal, err
		}
		return results, utils.NoError, nil
	}
	var results []types.GetCompetitionMetaData
	result, err := r.repo.GetCompetitionMetaData(ctx, *leagueId)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Error("Invalid league Id", "error", err)
			return results, utils.ErrNotFound, errors.New("Invalid league Id, please select an existing league Id")
		}

		slog.Error("Error getting competition meta data", "error", err)
		return results, utils.ErrInternal, err
	}

	return append(results, result), utils.NoError, nil
}

func (r *footballRepo) GetCompetitionTeamStandings(ctx context.Context, leagueId, seasonId *uuid.UUID) ([]types.StandingsQueryResponse, utils.ErrorType, error) {
	result, err := r.repo.GetCompetitionTeamStandings(ctx, leagueId, seasonId)

	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Error("Invalid league or season Id", "error", err)
			return result, utils.ErrNotFound, errors.New("Invalid league or season Id . Please select an existing league or season Id")
		}
		slog.Error("Error getting competition team standings", "error", err)
		return result, utils.ErrInternal, err
	}

	return result, utils.NoError, nil
}
