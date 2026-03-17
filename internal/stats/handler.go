package stats

import (
	"link-generator/configs"
	authsession "link-generator/internal/auth_session"
	"link-generator/internal/models"
	"link-generator/pkg/errorType"
	"link-generator/pkg/limiter"
	"link-generator/pkg/middleware"
	"link-generator/pkg/redis"
	"link-generator/pkg/response"
	"net/http"
	"strconv"
	"time"
)

type ILinkGetter interface {
	GetByID(id uint) (*models.Link, error)
}

type StatsHandlerDeps struct {
	*configs.Config
	AuthSessionService *authsession.AuthSessionService
	StatsRepository    *StatsRepository
	StatsService       *StatsService
	Redis              *redis.Redis
	RateLimiter        *limiter.LimiterService
	LinkRepository     ILinkGetter
}

type StatsHandler struct {
	*configs.Config
	responsePkg     response.Response
	StatsRepository *StatsRepository
	statsService    *StatsService
	redis           *redis.Redis
	linkRepository  ILinkGetter
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
		linkRepository:  deps.LinkRepository,
	}

	// Middlewares
	createMiddleware := middleware.Chain(
		middleware.IsAuthed(*deps.AuthSessionService),
		middleware.RateLimit(deps.RateLimiter, limiter.KeyByAccountID),
	)

	router.Handle("GET /stats", createMiddleware(handler.getStats()))
	router.Handle("GET /stats/link/{id}", createMiddleware(handler.getGroupedStatsByDate()))
}

func (stats *StatsHandler) getStats() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		queries := parsedTimeQuery([]string{"from", "to"}, req)

		var linkID *uint
		if raw := req.URL.Query().Get("linkId"); raw != "" {
			if parsed, err := strconv.ParseUint(raw, 10, 64); err == nil {
				id := uint(parsed)
				linkID = &id
			}
		}

		var cacheID uint
		if linkID != nil {
			cacheID = *linkID
		}
		cachedStats, _ := GetCachedStat[[]Stats](stats.redis, queries, strconv.Itoa(int(cacheID)))

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
		parsed, err := strconv.ParseUint(req.PathValue("id"), 10, 64)
		if err != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   &errorType.ErrorType{Error: "linkId must be a number"},
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		linkID := uint(parsed)
		queries := parsedTimeQuery([]string{"from", "to"}, req)

		link, linkErr := stats.linkRepository.GetByID(linkID)
		if linkErr != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   &errorType.ErrorType{Error: linkErr.Error()},
				Code:   http.StatusInternalServerError,
				Writer: w,
				Reader: req,
			})
			return
		}

		cachedStats, _ := GetCachedStat[[]GetStatByLink](stats.redis, queries, link.Hash)

		if cachedStats != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   cachedStats,
				Code:   http.StatusOK,
				Writer: w,
				Reader: req,
			})
			return
		}

		result, err := stats.StatsRepository.GetStatByLink(&StatsQuery{from: queries["from"], to: queries["to"], linkID: &linkID})

		if err != nil {
			stats.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		SetCachedStat(stats.redis, result, queries, link.Hash)

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
