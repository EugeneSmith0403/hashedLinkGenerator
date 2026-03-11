package user

import (
	"net/http"

	authsession "link-generator/internal/auth_session"
	"link-generator/internal/models"
	errorType "link-generator/pkg/errorType"
	"link-generator/pkg/helpers"
	"link-generator/pkg/middleware"
	"link-generator/pkg/request"
	"link-generator/pkg/response"
)

type UserMeResponse struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	Email        string                   `json:"email"`
	Is2FAEnabled bool                     `json:"is2faEnabled"`
	Subscription *models.SubscriptionInfo `json:"subscription"`
}

type Setup2FAResponse struct {
	QRCode string `json:"qrCode"`
}

type UserHandlerDeps struct {
	UserRepository      *UserRepository
	AccountService      models.IAccountService
	SubscriptionService models.ISubscriptionService
	AuthService         models.IAuthService
	AuthSeseionService  *authsession.AuthSessionService
}

type UserHandler struct {
	userRepository      *UserRepository
	accountService      models.IAccountService
	subscriptionService models.ISubscriptionService
	authService         models.IAuthService
	authSeseionService  *authsession.AuthSessionService
	responsePkg         response.Response
}

func NewUserHandlers(router *http.ServeMux, deps UserHandlerDeps) {
	handler := &UserHandler{
		userRepository:      deps.UserRepository,
		accountService:      deps.AccountService,
		subscriptionService: deps.SubscriptionService,
		authService:         deps.AuthService,
		authSeseionService:  deps.AuthSeseionService,
		responsePkg: *response.NewResponse(&response.ResponseOptions{
			HeadersMap: map[string]string{"Content-Type": "application/json"},
		}),
	}

	authMiddleware := middleware.Chain(
		middleware.IsAuthed(*deps.AuthSeseionService),
	)

	router.Handle("GET /users/me", authMiddleware(handler.handleGetMe()))
	router.Handle("POST /users/me/2fa/setup", authMiddleware(handler.handleSetup2FA()))
	router.Handle("POST /users/2fa/verify", handler.handler2FaVerify())
}

func (h *UserHandler) handleGetMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "unauthorized"},
				Writer: w,
				Reader: r,
				Code:   http.StatusUnauthorized,
			})
			return
		}

		foundUser, err := h.userRepository.FindByEmailWithAccounts(email)
		if err != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: err.Error()},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}
		if foundUser == nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "user not found"},
				Writer: w,
				Reader: r,
				Code:   http.StatusNotFound,
			})
			return
		}

		if len(foundUser.Accounts) == 0 {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "Account not found"},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}
		accInfo := foundUser.Accounts[0]

		resp := &UserMeResponse{
			ID:           foundUser.Model.ID,
			Name:         foundUser.Name,
			Email:        foundUser.Email,
			Is2FAEnabled: accInfo.Is2FAEnabled,
		}

		sub, _ := h.subscriptionService.GetSubscriptionByUserID(foundUser.Model.ID)
		if sub != nil {
			resp.Subscription = sub
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   resp,
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}

func (h *UserHandler) handleSetup2FA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "unauthorized"},
				Writer: w,
				Reader: r,
				Code:   http.StatusUnauthorized,
			})
			return
		}

		qrCode, setupErr := h.accountService.Setup2FA(email)
		if setupErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: setupErr.Error()},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   Setup2FAResponse{QRCode: qrCode},
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}

func (h *UserHandler) handler2FaVerify() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := request.HandleBody[VerifyRequest](req, w, h.responsePkg)
		if err != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   err,
				Writer: w,
				Reader: req,
				Code:   http.StatusBadRequest,
			})
			return
		}

		isValid := h.accountService.Verify2Fa(body.Code, body.Email)
		if !isValid {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "invalid code"},
				Writer: w,
				Reader: req,
				Code:   http.StatusBadRequest,
			})
			return
		}

		token, expTime, tokenErr := h.authService.GenerateToken(body.Email)
		if tokenErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: tokenErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		foundUser, err := h.userRepository.FindByEmailWithAccounts(body.Email)
		if err != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: err.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}
		if foundUser == nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "user not found"},
				Writer: w,
				Reader: req,
				Code:   http.StatusNotFound,
			})
			return
		}

		if len(foundUser.Accounts) == 0 {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "Account not found"},
				Writer: w,
				Reader: req,
				Code:   http.StatusNotFound,
			})
			return
		}

		options := &authsession.AddOptions{
			AccountID: foundUser.Accounts[0].ID,
			Token:     token,
			IpAddress: helpers.GetClientIP(req),
			UserAgent: req.UserAgent(),
			IsVerify:  isValid,
			ExpiresAt: expTime,
		}

		_, authSessErr := h.authSeseionService.Update(options)

		if authSessErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: authSessErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusNotFound,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   VerifyResponse{Token: token},
			Writer: w,
			Reader: req,
			Code:   http.StatusOK,
		})
	}
}
