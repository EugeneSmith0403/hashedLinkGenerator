package subscription

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"link-generator/internal/account"
	internalJWT "link-generator/internal/jwt"
	"link-generator/internal/payments/plan"
	errorType "link-generator/pkg/errorType"
	"link-generator/pkg/middleware"
	"link-generator/pkg/request"
	"link-generator/pkg/response"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type SubscriptionResponse struct {
	ID                 uint               `json:"id"`
	CreatedAt          time.Time          `json:"createdAt"`
	UserID             uint               `json:"userId"`
	PlanID             uint               `json:"planId"`
	Status             SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time          `json:"currentPeriodStart"`
	CurrentPeriodEnd   time.Time          `json:"currentPeriodEnd"`
	CancelAt           *time.Time         `json:"cancelAt"`
	CanceledAt         *time.Time         `json:"canceledAt"`
	TrialStart         *time.Time         `json:"trialStart"`
	TrialEnd           *time.Time         `json:"trialEnd"`
	IsPaymentIntent    bool               `json:"isPaymentIntent"`
}

func toSubscriptionResponse(s *Subscription) *SubscriptionResponse {
	return &SubscriptionResponse{
		ID:                 s.ID,
		CreatedAt:          s.CreatedAt,
		UserID:             s.UserID,
		PlanID:             s.PlanID,
		Status:             s.Status,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		CancelAt:           s.CancelAt,
		CanceledAt:         s.CanceledAt,
		TrialStart:         s.TrialStart,
		TrialEnd:           s.TrialEnd,
		IsPaymentIntent:    strings.HasPrefix(s.BillingID, "pi_"),
	}
}

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

	router.Handle("GET /subscriptions/me", authMiddleware(handler.handleGetCurrentSubscription()))
	router.Handle("POST /subscriptions/method", authMiddleware(handler.handleAddPaymentMethod()))
	router.Handle("POST /subscriptions", authMiddleware(handler.handleCreateSubscription()))
	router.Handle("PATCH /subscriptions/cancel", authMiddleware(handler.handleCancelSubscription()))
}

const (
	errUnauthorized      = "unauthorized"
	errPlanNotFound      = "plan not found"
	errPlanNoStripePrice = "plan has no stripe price configured"
	errActiveSubNotFound = "active subscription not found"
	errNoPaymentMethod   = "try again"
)

func stripeSubErrMsg(err error) string {
	var stripeErr *stripeGo.Error
	if errors.As(err, &stripeErr) {
		return stripeErr.Msg
	}
	return err.Error()
}

func (h *SubscriptionHandler) handleGetCurrentSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: errUnauthorized},
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

		sub, err := h.subscriptionService.GetSubscriptionByUserId(foundAccount.UserID)
		if err != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: err.Error()},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		if sub == nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   nil,
				Writer: w,
				Reader: r,
				Code:   http.StatusNoContent,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   toSubscriptionResponse(sub),
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}

func (h *SubscriptionHandler) handleAddPaymentMethod() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: errUnauthorized},
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
				Data:   errorType.ErrorType{Error: errUnauthorized},
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
				Data:   errorType.ErrorType{Error: errPlanNotFound},
				Writer: w,
				Reader: r,
				Code:   http.StatusNotFound,
			})
			return
		}

		if curPlan.StripePriceID == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: errPlanNoStripePrice},
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
			code := http.StatusInternalServerError
			msg := stripeSubErrMsg(subErr)
			var stripeErr *stripeGo.Error
			if errors.As(subErr, &stripeErr) && stripeErr.Code == stripeGo.ErrorCodeResourceMissing {
				code = http.StatusUnprocessableEntity
				msg = errNoPaymentMethod
			}
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: msg},
				Writer: w,
				Reader: r,
				Code:   code,
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

func (h *SubscriptionHandler) handleCancelSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		email, ok := r.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: errUnauthorized},
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

		sub, subErr := h.subscriptionService.GetSubscriptionByUserId(foundAccount.UserID)
		if subErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: subErr.Error()},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}
		if sub == nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: errActiveSubNotFound},
				Writer: w,
				Reader: r,
				Code:   http.StatusNotFound,
			})
			return
		}

		canceled, cancelErr := h.subscriptionService.CancelSubscription(sub.BillingID)
		if cancelErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeSubErrMsg(cancelErr)},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   canceled,
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}
