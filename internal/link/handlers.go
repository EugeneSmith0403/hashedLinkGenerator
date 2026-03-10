package link

import (
	"link-generator/configs"
	"link-generator/internal/jwt"
	"link-generator/internal/models"
	"link-generator/internal/user"
	"link-generator/pkg/event"
	"link-generator/pkg/middleware"
	"link-generator/pkg/request"
	"link-generator/pkg/response"
	"fmt"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type LinkHandlerDeps struct {
	*configs.Config
	LinkRepository      *LinkRepository
	UserRepository      *user.UserRepository
	JWTService          *jwt.JWTService
	EventBus            *event.EventBus
	SubscriptionService middleware.SubChecker
}

type LinkHandler struct {
	*configs.Config
	responsePkg    response.Response
	LinkRepository *LinkRepository
	UserRepository *user.UserRepository
	EventBus       *event.EventBus
}

func NewLinkHandler(router *http.ServeMux, deps LinkHandlerDeps) {
	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}
	handler := &LinkHandler{
		Config:         deps.Config,
		responsePkg:    *response.NewResponse(options),
		LinkRepository: deps.LinkRepository,
		UserRepository: deps.UserRepository,
		EventBus:       deps.EventBus,
	}

	// Middlewares
	authMiddleware := middleware.Chain(
		middleware.IsAuthed(deps.JWTService),
	)
	createMiddleware := middleware.Chain(
		middleware.IsAuthed(deps.JWTService),
		middleware.HasActiveSubscription(deps.UserRepository, deps.SubscriptionService),
	)

	router.HandleFunc("GET /{hash}", handler.GetTo())
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

		go link.EventBus.Publish(event.Event{
			Type: event.LinkVisitedEVent,
			Data: int(result.Model.ID),
		})

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
