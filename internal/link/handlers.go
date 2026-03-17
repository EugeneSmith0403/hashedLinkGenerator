package link

import (
	"fmt"
	"link-generator/configs"
	authsession "link-generator/internal/auth_session"
	"link-generator/internal/models"
	"link-generator/internal/publishers"
	"link-generator/internal/stats"
	"link-generator/internal/user"
	"link-generator/pkg/event"
	"link-generator/pkg/limiter"
	"link-generator/pkg/middleware"
	rabbitmq "link-generator/pkg/rabbitMq"
	"link-generator/pkg/request"
	"link-generator/pkg/response"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type LinkHandlerDeps struct {
	*configs.Config
	LinkRepository      *LinkRepository
	UserRepository      *user.UserRepository
	AuthSessionService  *authsession.AuthSessionService
	EventBus            *event.EventBus
	SubscriptionService middleware.SubChecker
	RabbitMq            *rabbitmq.RabbitMq
	StatsService        *stats.StatsService
	RateLimiter         *limiter.LimiterService
	IPRateLimiter       *limiter.LimiterService
}

type LinkHandler struct {
	*configs.Config
	responsePkg    response.Response
	LinkRepository *LinkRepository
	UserRepository *user.UserRepository
	EventBus       *event.EventBus
	StatsPublisher *publishers.StatsPublisher
	StatsService   *stats.StatsService
}

func NewLinkHandler(router *http.ServeMux, deps LinkHandlerDeps) {
	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	statsPub := publishers.NewStatsPublisher(deps.RabbitMq)
	statsPub.CreateExchangeAndQueue()

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}
	handler := &LinkHandler{
		Config:         deps.Config,
		responsePkg:    *response.NewResponse(options),
		LinkRepository: deps.LinkRepository,
		UserRepository: deps.UserRepository,
		EventBus:       deps.EventBus,
		StatsPublisher: statsPub,
		StatsService:   deps.StatsService,
	}

	// Middlewares
	authMiddleware := middleware.Chain(
		middleware.IsAuthed(*deps.AuthSessionService),
		middleware.RateLimit(deps.RateLimiter, limiter.KeyByAccountID),
	)
	createMiddleware := middleware.Chain(
		middleware.IsAuthed(*deps.AuthSessionService),
		middleware.RateLimit(deps.RateLimiter, limiter.KeyByAccountID),
		middleware.HasActiveSubscription(
			func(email string) (uint, error) {
				u, err := deps.UserRepository.FindByEmail(email)
				if err != nil {
					return 0, err
				}
				if u == nil {
					return 0, fmt.Errorf("user not found")
				}
				return u.ID, nil
			},
			deps.SubscriptionService,
		),
	)

	router.Handle("GET /{hash}", middleware.RateLimit(deps.IPRateLimiter, limiter.KeyByIP)(handler.GetTo()))
	router.Handle("GET /links", authMiddleware(handler.List()))
	router.Handle("POST /link", createMiddleware(handler.Create()))
	router.Handle("PATCH /link/{id}", authMiddleware(handler.Update()))
	router.Handle("DELETE /link/{id}", authMiddleware(handler.Delete()))
}

func (link *LinkHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		email := req.Context().Value(middleware.ContextEmailKey).(string)
		currentUser, err := link.UserRepository.FindByEmail(email)
		if err != nil || currentUser == nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   "user not found",
				Code:   http.StatusUnauthorized,
				Writer: w,
				Reader: req,
			})
			return
		}

		links, err := link.LinkRepository.GetByUserID(currentUser.ID)
		if err != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusInternalServerError,
				Writer: w,
				Reader: req,
			})
			return
		}

		link.responsePkg.Json(&response.JsonOptions{
			Data:   links,
			Code:   http.StatusOK,
			Writer: w,
			Reader: req,
		})
	}
}

func (link *LinkHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := request.HandleBody[LinkCreateRequest](req, w, link.responsePkg)

		if err != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		email := req.Context().Value(middleware.ContextEmailKey).(string)
		currentUser, err := link.UserRepository.FindByEmail(email)
		if err != nil || currentUser == nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   "user not found",
				Code:   http.StatusUnauthorized,
				Writer: w,
				Reader: req,
			})
			return
		}

		createdLink := models.NewLink(body.Url, currentUser.ID)

		result, errLink := link.LinkRepository.Create(createdLink)

		if errLink != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   errLink,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		link.responsePkg.Json(&response.JsonOptions{
			Data:   result,
			Code:   http.StatusOK,
			Writer: w,
			Reader: req,
		})

	}
}

func (link *LinkHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		body, err := request.HandleBody[LinkUpdateResponse](req, w, link.responsePkg)

		em := req.Context().Value(middleware.ContextEmailKey).(string)

		fmt.Println(em)

		if err != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		idStr := req.PathValue("id")

		id, strErr := strconv.ParseUint(idStr, 10, 64)

		if err != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   strErr,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		updatedLink, updtErr := link.LinkRepository.Update(&models.Link{
			Model: gorm.Model{
				ID: uint(id),
			},
			Url:  body.Url,
			Hash: body.Hash,
		})

		if updtErr != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   updtErr,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		link.responsePkg.Json(&response.JsonOptions{
			Data:   updatedLink,
			Code:   http.StatusCreated,
			Writer: w,
			Reader: req,
		})
	}
}

func (link *LinkHandler) GetTo() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		hash := req.PathValue("hash")

		result, err := link.LinkRepository.GetByHash(hash)

		if err != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		clientContext := link.StatsService.BuildClientContext(req)

		data := &models.LinkTransitionWitHash{
			LinkTransition: &models.LinkTransition{
				LinkID:    int64(result.ID),
				ClickedAt: clientContext.Timestamp,

				IP:           clientContext.IP,
				ForwardedFor: clientContext.ForwardedFor,
				RealIP:       clientContext.RealIP,
				RemoteAddr:   clientContext.RemoteAddr,
				RemotePort:   clientContext.RemotePort,
				Country:      clientContext.Country,

				UserAgent:      clientContext.UserAgent,
				Accept:         clientContext.Accept,
				AcceptLanguage: clientContext.AcceptLanguage,
				AcceptEncoding: clientContext.AcceptEncoding,
				Origin:         clientContext.Origin,
				Referer:        clientContext.Referer,

				Fingerprint:    clientContext.Fingerprint,
				RequestID:      clientContext.RequestID,
				ForwardedProto: clientContext.ForwardedProto,
				ForwardedHost:  clientContext.ForwardedHost,
				ForwardedPort:  clientContext.ForwardedPort,
				Scheme:         clientContext.Scheme,
			},
			FilterHash: hash,
		}

		go link.StatsPublisher.PublishToQueue(models.StatsLinkVisited, data)

		http.Redirect(w, req, result.Url, http.StatusTemporaryRedirect)
	}
}

func (link *LinkHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		idStr := req.PathValue("id")

		id, strErr := strconv.ParseUint(idStr, 10, 64)

		el, err := link.LinkRepository.getById(uint(id))

		if err != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   strErr,
				Code:   http.StatusInternalServerError,
				Writer: w,
				Reader: req,
			})
			return
		}

		if el == nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   strErr,
				Code:   http.StatusNotFound,
				Writer: w,
				Reader: req,
			})
			return
		}

		res, strErr := link.LinkRepository.Delete(&models.Link{Model: gorm.Model{ID: uint(id)}})

		if strErr != nil {
			link.responsePkg.Json(&response.JsonOptions{
				Data:   strErr,
				Code:   http.StatusBadRequest,
				Writer: w,
				Reader: req,
			})
			return
		}

		link.responsePkg.Json(&response.JsonOptions{
			Data:   res,
			Writer: w,
			Reader: req,
			Code:   http.StatusNoContent,
		})
	}
}
