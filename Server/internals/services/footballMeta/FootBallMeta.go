package footballmeta

import (
	"context"
	"errors"
	"log/slog"
	"sort"

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
	GetAllSeasons(ctx context.Context) ([]types.GetSeason, []types.GetLeagueSeason, utils.ErrorType, error)
	GetKnockoutData(ctx context.Context, leagueId, seasonId *uuid.UUID) ([]types.KnockoutStageData, utils.ErrorType, error)
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

var knockoutStageOrder = map[string]int{
	"KNOCKOUT_ROUND_PLAY_OFFS": 0,
	"LAST_16":                  1,
	"QUARTER_FINALS":           2,
	"SEMI_FINALS":              3,
	"FINAL":                    4,
}

func (r *footballRepo) GetKnockoutData(ctx context.Context, leagueId, seasonId *uuid.UUID) ([]types.KnockoutStageData, utils.ErrorType, error) {
	rows, err := r.repo.GetKnockoutMatches(ctx, leagueId, seasonId)
	if err != nil {
		slog.Error("Error getting knockout matches", "error", err)
		return nil, utils.ErrInternal, err
	}

	// Group rows by stage, then by normalized team-pair key.
	type pairKey struct{ a, b string }
	normalizeKey := func(id1, id2 uuid.UUID) pairKey {
		s1, s2 := id1.String(), id2.String()
		if s1 < s2 {
			return pairKey{s1, s2}
		}
		return pairKey{s2, s1}
	}

	// stageRows preserves insertion order for stages.
	stageOrder := []string{}
	seenStages := map[string]bool{}
	stagePairs := map[string]map[pairKey][]types.KnockoutMatchRow{}

	for _, row := range rows {
		if !seenStages[row.Stage] {
			seenStages[row.Stage] = true
			stageOrder = append(stageOrder, row.Stage)
			stagePairs[row.Stage] = map[pairKey][]types.KnockoutMatchRow{}
		}
		k := normalizeKey(row.HomeTeamID, row.AwayTeamID)
		stagePairs[row.Stage][k] = append(stagePairs[row.Stage][k], row)
	}

	// Sort stages by the canonical UCL order.
	sort.Slice(stageOrder, func(i, j int) bool {
		oi, ok1 := knockoutStageOrder[stageOrder[i]]
		oj, ok2 := knockoutStageOrder[stageOrder[j]]
		if !ok1 {
			oi = 99
		}
		if !ok2 {
			oj = 99
		}
		return oi < oj
	})

	result := make([]types.KnockoutStageData, 0, len(stageOrder))
	for _, stage := range stageOrder {
		matchups := []types.KnockoutMatchup{}
		for _, matchRows := range stagePairs[stage] {
			matchups = append(matchups, buildKnockoutMatchup(matchRows))
		}
		result = append(result, types.KnockoutStageData{Stage: stage, Matchups: matchups})
	}

	return result, utils.NoError, nil
}

// buildKnockoutMatchup converts 1–2 raw match rows into a structured matchup.
// The first row (by created_at ASC) is treated as leg1; its home team is team1.
func buildKnockoutMatchup(rows []types.KnockoutMatchRow) types.KnockoutMatchup {
	if len(rows) == 0 {
		return types.KnockoutMatchup{}
	}

	first := rows[0]
	team1 := types.KnockoutTeamInfo{
		ID:        first.HomeTeamID,
		Name:      first.HomeTeamName,
		ShortName: first.HomeTeamShortName,
		TLA:       first.HomeTeamTLA,
		Emblem:    first.HomeTeamEmblem,
	}
	team2 := types.KnockoutTeamInfo{
		ID:        first.AwayTeamID,
		Name:      first.AwayTeamName,
		ShortName: first.AwayTeamShortName,
		TLA:       first.AwayTeamTLA,
		Emblem:    first.AwayTeamEmblem,
	}

	mu := types.KnockoutMatchup{Team1: team1, Team2: team2}

	for _, row := range rows {
		leg := &types.KnockoutLeg{
			HomeTeamID: row.HomeTeamID,
			HomeScore:  row.HomeScore,
			AwayScore:  row.AwayScore,
			Status:     row.Status,
		}
		if row.HomeTeamID == team1.ID {
			mu.Leg1 = leg
		} else {
			mu.Leg2 = leg
		}
	}

	// Compute aggregate when both legs have final scores.
	if mu.Leg1 != nil && mu.Leg1.HomeScore != nil && mu.Leg1.AwayScore != nil &&
		mu.Leg2 != nil && mu.Leg2.HomeScore != nil && mu.Leg2.AwayScore != nil {
		// team1 goals = leg1 home + leg2 away
		// team2 goals = leg1 away + leg2 home
		t1 := *mu.Leg1.HomeScore + *mu.Leg2.AwayScore
		t2 := *mu.Leg1.AwayScore + *mu.Leg2.HomeScore
		mu.Team1AggGoals = &t1
		mu.Team2AggGoals = &t2
		if t1 > t2 {
			id := team1.ID
			mu.WinnerTeamID = &id
		} else if t2 > t1 {
			id := team2.ID
			mu.WinnerTeamID = &id
		}
	} else if mu.Leg1 != nil && mu.Leg1.HomeScore != nil && mu.Leg1.AwayScore != nil && mu.Leg2 == nil {
		// Single-leg final: aggregate = leg1 scores
		t1 := *mu.Leg1.HomeScore
		t2 := *mu.Leg1.AwayScore
		mu.Team1AggGoals = &t1
		mu.Team2AggGoals = &t2
		if t1 > t2 {
			id := team1.ID
			mu.WinnerTeamID = &id
		} else if t2 > t1 {
			id := team2.ID
			mu.WinnerTeamID = &id
		}
	}

	return mu
}

func (r *footballRepo) GetAllSeasons(ctx context.Context) ([]types.GetSeason, []types.GetLeagueSeason, utils.ErrorType, error) {
	var allSeasons []types.GetSeason
	var allLeagueSeasons []types.GetLeagueSeason
	allSeasons, err := r.repo.GetAllSeasons(ctx)

	if err != nil {
		slog.Error("Error in getting all season service", "error", err)
		return allSeasons, allLeagueSeasons, utils.ErrInternal, err
	}
	allLeagueSeasons, err = r.repo.GetAllLeagueSeasons(ctx)

	if err != nil {
		slog.Error("Error in getting all league season service ", "error", err)
		return allSeasons, allLeagueSeasons, utils.ErrInternal, err
	}

	return allSeasons, allLeagueSeasons, utils.NoError, nil
}
