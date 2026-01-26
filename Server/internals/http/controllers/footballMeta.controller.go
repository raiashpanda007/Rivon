package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	footballmeta "github.com/raiashpanda007/rivon/internals/services/footballMeta"
	"github.com/raiashpanda007/rivon/internals/types"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type FootballMetaController interface {
	GetCompetitions(res http.ResponseWriter, req *http.Request)
	GetCompetitionTeamStandings(res http.ResponseWriter, req *http.Request)
}
type footballRepoSvc struct {
	svc footballmeta.FootballMetaServices
}

func InitFootballMetaController(db *pgxpool.Pool) FootballMetaController {
	services := footballmeta.InitFootBallMetaServices(db)
	return &footballRepoSvc{
		svc: services,
	}

}

func (r *footballRepoSvc) GetCompetitions(res http.ResponseWriter, req *http.Request) {

	leagueIdStr := req.URL.Query().Get("leagueId")
	var leagueId *uuid.UUID = nil
	if leagueIdStr != "" {
		id, err := uuid.Parse(leagueIdStr)
		if err != nil {
			slog.Error("Invalid league Id", "error", err)
			utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, errors.New("Invalid league Id")))
			return
		}
		leagueId = &id
	}
	result, errType, err := r.svc.GetCompetitions(req.Context(), leagueId)
	if err != nil {
		slog.Error("Error getting competitions", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}
	utils.WriteJson(res, http.StatusOK, utils.Response[[]types.GetCompetitionMetaData]{
		Status:  200,
		Heading: "Competition Details",
		Message: "Competition meta data",
		Data:    result,
	})
}

func (r *footballRepoSvc) GetCompetitionTeamStandings(res http.ResponseWriter, req *http.Request) {
	leagueIdStr := req.URL.Query().Get("leagueId")
	seasonIdStr := req.URL.Query().Get("seasonId")

	var leagueId *uuid.UUID = nil
	var seasonId *uuid.UUID = nil

	if leagueIdStr != "" {
		id, err := uuid.Parse(leagueIdStr)
		if err != nil {
			slog.Error("Invalid league Id", "error", err)
			utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, errors.New("Invalid league Id")))
			return
		}
		leagueId = &id
	}

	if seasonIdStr != "" {
		id, err := uuid.Parse(seasonIdStr)
		if err != nil {
			slog.Error("Invalid season Id", "error", err)
			utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, errors.New("Invalid season Id")))
			return
		}
		seasonId = &id
	}

	leagueUUID := leagueId

	seasonUUID := seasonId

	result, errType, err := r.svc.GetCompetitionTeamStandings(req.Context(), leagueUUID, seasonUUID)

	if err != nil {
		slog.Error("Error getting competition team standings", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}

	utils.WriteJson(res, http.StatusOK, utils.Response[[]types.StandingsQueryResponse]{
		Heading: "OK",
		Message: "Standings Fetched",
		Data:    result,
		Status:  200,
	})

}
