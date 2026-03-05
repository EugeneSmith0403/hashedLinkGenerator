package payment

import (
	"errors"
	"net/http"

	"adv/go-http/internal/account"
	internalJWT "adv/go-http/internal/jwt"
	errorType "adv/go-http/pkg/errorType"
	"adv/go-http/pkg/middleware"
	"adv/go-http/pkg/response"
)

type PaymentHandlerDeps struct {
	PaymentRepository *PaymentRepository
	JWTService        *internalJWT.JWTService
	AccountService    *account.AccountService
}

type PaymentHandler struct {
	paymentRepository *PaymentRepository
	responsePkg       response.Response
	accountService    *account.AccountService
}

func NewPaymentHandler(router *http.ServeMux, deps PaymentHandlerDeps) {
	handler := &PaymentHandler{
		paymentRepository: deps.PaymentRepository,
		responsePkg: *response.NewResponse(&response.ResponseOptions{
			HeadersMap: map[string]string{"Content-Type": "application/json"},
		}),
		accountService: deps.AccountService,
	}

	authMiddleware := middleware.Chain(
		middleware.IsAuthed(deps.JWTService),
	)

	router.Handle("GET /payments", authMiddleware(handler.handleGetPayments()))
}

func (h *PaymentHandler) handleGetPayments() http.HandlerFunc {
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

		foundAccount, err := h.accountService.GetAccountByEmail(email)
		if err != nil || foundAccount == nil {
			code := http.StatusInternalServerError
			if errors.Is(err, account.ErrUserNotFound) || errors.Is(err, account.ErrAccountNotFound) {
				code = http.StatusNotFound
			}
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: err.Error()},
				Writer: w,
				Reader: r,
				Code:   code,
			})
			return
		}

		payments, err := h.paymentRepository.GetByAccountID(foundAccount.ID)
		if err != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: err.Error()},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   payments,
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}
