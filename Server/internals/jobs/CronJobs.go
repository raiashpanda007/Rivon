package jobs

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
	"github.com/raiashpanda007/rivon/internals/registry"
	"github.com/raiashpanda007/rivon/internals/services/markets"
	"github.com/raiashpanda007/rivon/internals/types"
	"github.com/raiashpanda007/rivon/internals/utils"
	"golang.org/x/sync/errgroup"
)

type CronJobs interface {
	UpdateLeagueStats(ctx context.Context) error
}

type cronJobs struct {
	repo      FootBallMetaRepo
	cfg       *config.Config
	marketSvc markets.MarketServices
}

func timeConvertor(rawDate string) (time.Time, error) {
	return time.Parse("2006-01-02", rawDate)
}

func NewCronJobs(db *pgxpool.Pool, cfg *config.Config, orderRedis *redis.Client) CronJobs {
	repo := NewFootBallMetaRepo(db)
	marketServices := markets.NewMarketServices(db, orderRedis, registry.New())
	return &cronJobs{
		repo:      repo,
		cfg:       cfg,
		marketSvc: marketServices,
	}
}

func (r *cronJobs) UpdateLeagueStats(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	leagues, err := r.repo.GetCompetitions(ctx)
	if err != nil {
		return err
	}

	for _, league := range leagues {
		league := league
		g.Go(func() error {
			resp, err := utils.FootBallOrgAPICaller[struct{}, types.StandingsResponse](
				utils.ApiCallerProps[struct{}]{
					BaseURL: "https://api.football-data.org",
					Paths:   []string{"v4", "competitions", strconv.Itoa(league.FootballOrgId), "standings"},
					ReqType: utils.REQ_GET,
				},
				&r.cfg.FootbalOrgApiKey[0],
				r.cfg.FootbalOrgApiKey,
			)
			if err != nil {
				return err
			}

			startDate, err := timeConvertor(resp.Season.StartDate)
			if err != nil {
				return err
			}

			endDate, err := timeConvertor(resp.Season.EndDate)
			if err != nil {
				return err
			}

			seasonId, err := r.repo.SaveSeasons(
				ctx,
				resp.Filters.Season+"-"+league.Code,
				resp.Season.ID,
				startDate,
				endDate,
				resp.Season.CurrentMatchday,
				nil,
				league.Id,
			)
			if err != nil {
				return err
			}

			for _, standing := range resp.Standings {
				if standing.Type != "TOTAL" {
					continue
				}

				for _, tableRow := range standing.Table {
					idStr := strconv.Itoa(tableRow.Team.ID)
					tla := tableRow.Team.TLA
					if tla == "" {
						tla = idStr
					}
					uniqueTla := tla + "_" + idStr

					teamID, err := r.repo.SaveTeamWithLeague(
						ctx,
						tableRow.Team.Name,
						tableRow.Team.ShortName,
						uniqueTla,
						uniqueTla,
						tableRow.Team.Crest,
						tableRow.Team.ID,
						league.Id,
						seasonId,
					)
					if err != nil {
						return err
					}

					_, _, err = r.marketSvc.CreateMarket(ctx, teamID.String(), tableRow.Team.Name, tableRow.Team.TLA, 0, 0, 0, 0)

					if err != nil {
						return err
					}

					if err := r.repo.SaveStandings(
						ctx,
						*teamID,
						league.Id,
						seasonId,
						tableRow.PlayedGames,
						tableRow.Won,
						tableRow.Lost,
						tableRow.Draw,
						tableRow.GoalsFor,
						tableRow.GoalsAgainst,
						tableRow.Position,
					); err != nil {
						return err
					}
				}
			}

			// For competitions with knockout rounds (e.g. CL), fetch bracket data.
			if league.Code == "CL" {
				for _, stage := range knockoutStages {
					if err := r.saveKnockoutStage(ctx, league, seasonId, stage); err != nil {
						slog.Warn("Failed to save knockout stage", "league", league.Code, "stage", stage, "error", err)
					}
				}
			}

			return nil
		})
	}

	return g.Wait()
}

// knockoutStages are the stage identifiers used by football-data.org for UCL knockouts.
var knockoutStages = []string{
	"KNOCKOUT_ROUND_PLAY_OFFS",
	"LAST_16",
	"QUARTER_FINALS",
	"SEMI_FINALS",
	"FINAL",
}

func (r *cronJobs) saveKnockoutStage(ctx context.Context, league types.LeagueStruct, seasonId uuid.UUID, stage string) error {
	resp, err := utils.FootBallOrgAPICaller[struct{}, types.MatchesResponse](
		utils.ApiCallerProps[struct{}]{
			BaseURL: "https://api.football-data.org",
			Paths:   []string{"v4", "competitions", strconv.Itoa(league.FootballOrgId), "matches"},
			Params:  map[string]string{"stage": stage},
			ReqType: utils.REQ_GET,
		},
		&r.cfg.FootbalOrgApiKey[0],
		r.cfg.FootbalOrgApiKey,
	)
	if err != nil {
		return err
	}

	for _, match := range resp.Matches {
		homeTeamId, err := r.repo.GetTeamByFootballOrgId(ctx, match.HomeTeam.ID)
		if err != nil {
			slog.Warn("Home team not found for knockout match", "teamOrgId", match.HomeTeam.ID, "matchId", match.ID)
			continue
		}
		awayTeamId, err := r.repo.GetTeamByFootballOrgId(ctx, match.AwayTeam.ID)
		if err != nil {
			slog.Warn("Away team not found for knockout match", "teamOrgId", match.AwayTeam.ID, "matchId", match.ID)
			continue
		}

		if err := r.repo.SaveKnockoutMatch(
			ctx,
			league.Id,
			seasonId,
			match.ID,
			match.Stage,
			*homeTeamId,
			*awayTeamId,
			match.Score.FullTime.Home,
			match.Score.FullTime.Away,
			match.Status,
		); err != nil {
			slog.Error("Error saving knockout match", "matchId", match.ID, "error", err)
		}
	}
	return nil
}
