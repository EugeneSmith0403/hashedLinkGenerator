package subscription

import (
	"errors"
	"net/http"

	"adv/go-http/internal/account"
	internalJWT "adv/go-http/internal/jwt"
	"adv/go-http/internal/payments/plan"
	errorType "adv/go-http/pkg/errorType"
	"adv/go-http/pkg/middleware"
	"adv/go-http/pkg/request"
	"adv/go-http/pkg/response"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type SubscriptionHandlerDeps struct {
	SubscriptionService *SubscriptionService
	JWTService          *internalJWT.JWTService
	AccountService      *account.AccountService
	PlanRepository      *plan.PlanRepository
}

type SubscriptionHandler struct {
	subscriptionService *SubscriptionService
	responsePkg         response.Response
	accountService      *account.AccountService
	planRepository      *plan.PlanRepository
}

func NewSubscriptionHandlers(router *http.ServeMux, deps SubscriptionHandlerDeps) {
	handler := &SubscriptionHandler{
		subscriptionService: deps.SubscriptionService,
		responsePkg: *response.NewResponse(&response.ResponseOptions{
			HeadersMap: map[string]string{"Content-Type": "application/json"},
		}),
		accountService: deps.AccountService,
		planRepository: deps.PlanRepository,
	}

	authMiddleware := middleware.Chain(
		middleware.IsAuthed(deps.JWTService),
	)

	router.Handle("POST /subscriptions/method", authMiddleware(handler.handleAddPaymentMethod()))
	router.Handle("POST /subscriptions", authMiddleware(handler.handleCreateSubscription()))
}

func stripeSubErrMsg(err error) string {
	var stripeErr *stripeGo.Error
	if errors.As(err, &stripeErr) {
		return stripeErr.Msg
	}
	return err.Error()
}

func (h *SubscriptionHandler) handleAddPaymentMethod() http.HandlerFunc {
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
				Data:   errorType.ErrorType{Error: stripeSubErrMsg(err)},
				Writer: w,
				Reader: r,
				Code:   code,
			})
			return
		}

		si, siErr := h.subscriptionService.AddPaymentMethod(foundAccount.CustomerID)
		if siErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeSubErrMsg(siErr)},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   SetupIntentResponse{ClientSecret: si.ClientSecret},
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}

func (h *SubscriptionHandler) handleCreateSubscription() http.HandlerFunc {
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
				Data:   errorType.ErrorType{Error: stripeSubErrMsg(err)},
				Writer: w,
				Reader: r,
				Code:   code,
			})
			return
		}

		body, bodyErr := request.HandleBody[CreateSubscriptionPayload](r, w, h.responsePkg)
		if bodyErr != nil {
			return
		}

		curPlan, planErr := h.planRepository.GetByID(body.PlanId)
		if planErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: planErr.Error()},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}
		if curPlan == nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "plan not found"},
				Writer: w,
				Reader: r,
				Code:   http.StatusNotFound,
			})
			return
		}

		if curPlan.StripePriceID == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "plan has no stripe price configured"},
				Writer: w,
				Reader: r,
				Code:   http.StatusUnprocessableEntity,
			})
			return
		}

		sub, subErr := h.subscriptionService.CreateSubscription(
			foundAccount.UserID,
			curPlan.ID,
			foundAccount.CustomerID,
			curPlan.StripePriceID,
		)
		if subErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeSubErrMsg(subErr)},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   sub,
			Writer: w,
			Reader: r,
			Code:   http.StatusCreated,
		})
	}
}
