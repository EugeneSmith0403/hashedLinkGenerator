package stats

import (
	"link-generator/configs"
	"link-generator/internal/jwt"
	"link-generator/pkg/errorType"
	"link-generator/pkg/middleware"
	"link-generator/pkg/redis"
	"link-generator/pkg/response"
	"net/http"
	"strconv"
	"time"
)

type StatsHandlerDeps struct {
	*configs.Config
	JWTService      *jwt.JWTService
	StatsRepository *StatsRepository
	StatsService    *StatsService
	Redis           *redis.Redis
}

type StatsHandler struct {
	*configs.Config
	responsePkg     response.Response
	StatsRepository *StatsRepository
	statsService    *StatsService
	redis           *redis.Redis
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
		redis:           deps.Redis,
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
		email, _ := req.Context().Value(middleware.ContextEmailKey).(string)
		queries := parsedTimeQuery([]string{"from", "to"}, req)

		var linkID *uint
		if raw := req.URL.Query().Get("linkId"); raw != "" {
			if parsed, err := strconv.ParseUint(raw, 10, 64); err == nil {
				id := uint(parsed)
				linkID = &id
			}
		}

		cachedStats, _ := GetCachedStat[[]Stats](stats.redis, queries, email)

		if cachedStats != nil && linkID == nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   cachedStats,
				Code:   http.StatusOK,
				Writer: w,
				Reader: req,
			})
			return
		}

		result, err := stats.StatsRepository.GetStats(&StatsQuery{from: queries["from"], to: queries["to"], linkID: linkID})

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
		email, ok := req.Context().Value(middleware.ContextEmailKey).(string)

		if !ok {
			stats.responsePkg.Json(&response.JsonOptions{
				Data: &errorType.ErrorType{
					Error: "Unathrorized",
				},
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
		}

		queries := parsedTimeQuery([]string{"from", "to"}, req)

		cachedStats, _ := GetCachedStat[[]StatsGroupByDate](stats.redis, queries, email)

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

		SetCachedStat(stats.redis, result, queries, email)

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
