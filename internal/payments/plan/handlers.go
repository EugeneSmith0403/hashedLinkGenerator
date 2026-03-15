package plan

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	authsession "link-generator/internal/auth_session"
	"link-generator/pkg/limiter"
	"link-generator/pkg/middleware"
	"link-generator/pkg/redis"
	"link-generator/pkg/response"
)

const plansCacheKey = "plans:active"
const plansCacheTTL = 30 * 24 * time.Hour

type PlanHandlerDeps struct {
	PlanRepository     *PlanRepository
	Redis              *redis.Redis
	AuthSessionService *authsession.AuthSessionService
	RateLimiter        *limiter.LimiterService
}

type PlanHandler struct {
	responsePkg    response.Response
	planRepository *PlanRepository
	redis          *redis.Redis
}

func NewPlanHandler(router *http.ServeMux, deps PlanHandlerDeps) {
	handler := &PlanHandler{
		responsePkg: *response.NewResponse(&response.ResponseOptions{
			HeadersMap: map[string]string{"Content-Type": "application/json"},
		}),
		planRepository: deps.PlanRepository,
		redis:          deps.Redis,
	}

	authMiddleware := middleware.Chain(
		middleware.IsAuthed(*deps.AuthSessionService),
		middleware.RateLimit(deps.RateLimiter, limiter.KeyByAccountID),
	)

	router.Handle("GET /plans", authMiddleware(handler.getPlans()))
}

func (h *PlanHandler) getPlans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cached := h.redis.Get(plansCacheKey); cached != "" {
			var plans []*Plan
			if err := json.Unmarshal([]byte(cached), &plans); err == nil {
				h.responsePkg.Json(&response.JsonOptions{
					Data:   plans,
					Code:   http.StatusOK,
					Writer: w,
					Reader: r,
				})
				return
			}
		}

		plans, err := h.planRepository.GetAll()
		if err != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusInternalServerError,
				Writer: w,
				Reader: r,
			})
			return
		}

		if data, err := json.Marshal(plans); err == nil {
			h.redis.Set(plansCacheKey, string(data), plansCacheTTL)
		} else {
			log.Printf("[plan] cache marshal error: %v", err)
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   plans,
			Code:   http.StatusOK,
			Writer: w,
			Reader: r,
		})
	}
}
