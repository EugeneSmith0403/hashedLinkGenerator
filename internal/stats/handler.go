package stats

import (
	"adv/go-http/configs"
	"adv/go-http/internal/jwt"
	"adv/go-http/pkg/middleware"
	"adv/go-http/pkg/response"
	"net/http"
	"time"
)

type StatsHandlerDeps struct {
	*configs.Config
	JWTService      *jwt.JWTService
	StatsRepository *StatsRepository
	StatsService    *StatsService
}

type StatsHandler struct {
	*configs.Config
	responsePkg     response.Response
	StatsRepository *StatsRepository
	statsService    *StatsService
}

func NewStatsHandler(router *http.ServeMux, deps StatsHandlerDeps) {
	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}
	handler := &StatsHandler{
		Config:          deps.Config,
		responsePkg:     *response.NewResponse(options),
		StatsRepository: deps.StatsRepository,
		statsService:    deps.StatsService,
	}

	// Middlewares
	createMiddleware := middleware.Chain(
		middleware.IsAuthed(deps.JWTService),
	)

	router.Handle("GET /stats", createMiddleware(handler.getStats()))
	router.Handle("GET /stats/clicks", createMiddleware(handler.getGroupedStatsByDate()))
}

func (stats *StatsHandler) getStats() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		queries := parsedTimeQuery([]string{"from", "to"}, req)

		cachedStats, _ := GetCachedStat[[]Stats](stats.statsService, queries)

		if cachedStats != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   cachedStats,
				Code:   http.StatusOK,
				Writer: w,
				Reader: req,
			})
			return
		}

		result, err := stats.StatsRepository.GetStats(&StatsQuery{from: queries["from"], to: queries["to"]})

		if err != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		stats.responsePkg.Json(&response.JsonOptions{
			Data:   result,
			Code:   http.StatusOK,
			Writer: w,
			Reader: req,
		})

	}
}

func (stats *StatsHandler) getGroupedStatsByDate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		queries := parsedTimeQuery([]string{"from", "to"}, req)

		cachedStats, _ := GetCachedStat[[]StatsGroupByDate](stats.statsService, queries)

		if cachedStats != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   cachedStats,
				Code:   http.StatusOK,
				Writer: w,
				Reader: req,
			})
			return
		}

		result, err := stats.StatsRepository.GetStatsGroupByDate(&StatsQuery{from: queries["from"], to: queries["to"]})

		if err != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		SetCachedStat(stats.statsService, result, queries)

		stats.responsePkg.Json(&response.JsonOptions{
			Data:   result,
			Code:   http.StatusOK,
			Writer: w,
			Reader: req,
		})

	}
}

func parsedTimeQuery(names []string, req *http.Request) map[string]time.Time {
	var result map[string]time.Time = make(map[string]time.Time)
	for _, name := range names {
		value := req.URL.Query().Get(name)

		if value != "" {
			parsedValue, err := time.Parse(time.DateOnly, value)
			if err != nil {
				continue
			}
			result[name] = parsedValue
		}
	}

	return result
}
