package account

import (
	"errors"
	"net/http"

	internalJWT "adv/go-http/internal/jwt"
	"adv/go-http/internal/user"
	errorType "adv/go-http/pkg/errorType"
	"adv/go-http/pkg/middleware"
	"adv/go-http/pkg/request"
	"adv/go-http/pkg/response"
)

type AccountHandlerDeps struct {
	AccountService *AccountService
	UserRepository *user.UserRepository
	JWTService     *internalJWT.JWTService
}

type AccountHandler struct {
	responsePkg    response.Response
	AccountService *AccountService
	UserRepository *user.UserRepository
}

func NewAccountHandler(router *http.ServeMux, deps AccountHandlerDeps) {
	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}

	handler := &AccountHandler{
		responsePkg:    *response.NewResponse(options),
		AccountService: deps.AccountService,
		UserRepository: deps.UserRepository,
	}

	// Middlewares
	createMiddleware := middleware.Chain(
		middleware.IsAuthed(deps.JWTService),
	)

	router.Handle("POST /account", createMiddleware(handler.Create()))
	router.Handle("PATCH /account", createMiddleware(handler.Update()))
}

func (h *AccountHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		email, ok := req.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "unauthorized"},
				Writer: w,
				Reader: req,
				Code:   http.StatusUnauthorized,
			})
			return
		}

		foundUser, err := h.UserRepository.FindByEmail(email)
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

		body, bodyErr := request.HandleBody[UpdateAccountRequest](req, w, h.responsePkg)
		if bodyErr != nil {
			return
		}

		account, updateErr := h.AccountService.UpdateAccount(foundUser.Model.ID, body.Name, body.Email)
		if updateErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: updateErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data: &UpdateAccountResponse{
				ID:            account.ID,
				AccountStatus: account.AccountStatus,
				Provider:      account.Provider,
			},
			Writer: w,
			Reader: req,
			Code:   http.StatusOK,
		})
	}
}

func (h *AccountHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		email, ok := req.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "unauthorized"},
				Writer: w,
				Reader: req,
				Code:   http.StatusUnauthorized,
			})
			return
		}

		getAccount, accErr := h.AccountService.GetAccountByEmail(email)
		if accErr != nil && !errors.Is(accErr, ErrAccountNotFound) {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: accErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		if getAccount != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data: &CreateAccountResponse{
					ID:            getAccount.ID,
					AccountStatus: getAccount.AccountStatus,
					Provider:      getAccount.Provider,
				},
				Writer: w,
				Reader: req,
				Code:   http.StatusCreated,
			})

			return
		}

		foundUser, err := h.UserRepository.FindByEmail(email)
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

		account, createErr := h.AccountService.CreateAccount(foundUser.Model.ID, foundUser.Name, foundUser.Email)
		if createErr != nil && !errors.Is(createErr, ErrUserNotFound) {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: createErr.Error()},
				Writer: w,
				Reader: req,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data: &CreateAccountResponse{
				ID:            account.ID,
				AccountStatus: account.AccountStatus,
				Provider:      account.Provider,
			},
			Writer: w,
			Reader: req,
			Code:   http.StatusCreated,
		})
	}
}
