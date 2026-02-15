package jobs

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/config"
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

func NewCronJobs(db *pgxpool.Pool, cfg *config.Config) CronJobs {
	repo := NewFootBallMetaRepo(db)
	marketServices := markets.NewMarketServices(db)
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

			return nil
		})
	}

	return g.Wait()
}
