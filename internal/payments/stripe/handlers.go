package stripe

import (
	"errors"
	"net/http"
	"strings"

	"adv/go-http/internal/account"
	internalJWT "adv/go-http/internal/jwt"
	"adv/go-http/internal/payments/plan"
	"adv/go-http/internal/payments/subscription"
	stripeServices "adv/go-http/internal/payments/stripe/services"
	errorType "adv/go-http/pkg/errorType"
	"adv/go-http/pkg/middleware"
	"adv/go-http/pkg/request"
	"adv/go-http/pkg/response"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type StripeHandlerDeps struct {
	PaymentService      *stripeServices.PaymentService
	JWTService          *internalJWT.JWTService
	AccountService      *account.AccountService
	PlanRepository      *plan.PlanRepository
	SubscriptionService *subscription.SubscriptionService
}

type StripeHandler struct {
	paymentService      *stripeServices.PaymentService
	responsePkg         response.Response
	accountService      *account.AccountService
	planRepository      *plan.PlanRepository
	subscriptionService *subscription.SubscriptionService
}

func NewStripeHandlers(router *http.ServeMux, deps StripeHandlerDeps) {
	handler := &StripeHandler{
		paymentService: deps.PaymentService,
		responsePkg: *response.NewResponse(&response.ResponseOptions{
			HeadersMap: map[string]string{"Content-Type": "application/json"},
		}),
		accountService:      deps.AccountService,
		planRepository:      deps.PlanRepository,
		subscriptionService: deps.SubscriptionService,
	}

	authMiddleware := middleware.Chain(
		middleware.IsAuthed(deps.JWTService),
	)

	router.Handle("POST /stripe/paymentIntent", authMiddleware(handler.handlePaymentIntent()))
	router.Handle("POST /stripe/paymentIntent/confirm", authMiddleware(handler.handleConfirm()))
	router.Handle("POST /stripe/paymentIntent/cancel", authMiddleware(handler.handleCancelPaymentIntent()))
}

func stripeErrMsg(err error) string {
	var stripeErr *stripeGo.Error
	if errors.As(err, &stripeErr) {
		return stripeErr.Msg
	}
	return err.Error()
}

func (h *StripeHandler) handlePaymentIntent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(middleware.ContextEmailKey).(string)
		if !ok || email == "" {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "Unauthorized"},
				Writer: w,
				Reader: r,
				Code:   http.StatusUnauthorized,
			})
			return
		}

		foundAccount, err := h.accountService.GetAccountByEmail(email)
		if err != nil {
			code := http.StatusInternalServerError
			if errors.Is(err, account.ErrUserNotFound) || errors.Is(err, account.ErrAccountNotFound) {
				code = http.StatusNotFound
			}
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeErrMsg(err)},
				Writer: w,
				Reader: r,
				Code:   code,
			})
			return
		}

		body, bodyErr := request.HandleBody[PaymentIntentPayload](r, w, h.responsePkg)
		if bodyErr != nil {
			return
		}

		curPlan, planErr := h.planRepository.GetByID(body.PlanId)

		if planErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeErrMsg(planErr)},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		pi, piErr := h.paymentService.CreatePaymentIntent(stripeServices.CreatePaymentIntentParams{
			AccountId:  foundAccount.ID,
			UserId:     foundAccount.UserID,
			CustomerID: foundAccount.CustomerID,
			CardType:   body.CardType,
			Currency:   stripeGo.Currency(curPlan.Currency),
			Amount:     int64(curPlan.Cost),
			PlanId:     body.PlanId,
		})

		if piErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeErrMsg(piErr)},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   pi,
			Writer: w,
			Reader: r,
			Code:   http.StatusCreated,
		})
	}
}

func (h *StripeHandler) handleConfirm() http.HandlerFunc {
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
				Data:   errorType.ErrorType{Error: stripeErrMsg(err)},
				Writer: w,
				Reader: r,
				Code:   code,
			})
			return
		}

		body, bodyErr := request.HandleBody[ConfirmPaymentIntentPayload](r, w, h.responsePkg)
		if bodyErr != nil {
			return
		}

		confirmedResponse, confirmErr := h.paymentService.ConfirmPaymentIntent(body.PaymentId)
		if confirmErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeErrMsg(confirmErr)},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		h.responsePkg.Json(&response.JsonOptions{
			Data:   confirmedResponse,
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}

func (h *StripeHandler) handleCancelPaymentIntent() http.HandlerFunc {
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
				Data:   errorType.ErrorType{Error: stripeErrMsg(err)},
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
		if sub == nil || !strings.HasPrefix(sub.BillingID, "pi_") {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: "no active payment intent subscription found"},
				Writer: w,
				Reader: r,
				Code:   http.StatusNotFound,
			})
			return
		}

		if refundErr := h.paymentService.RefundPaymentIntent(sub.BillingID); refundErr != nil {
			h.responsePkg.Json(&response.JsonOptions{
				Data:   errorType.ErrorType{Error: stripeErrMsg(refundErr)},
				Writer: w,
				Reader: r,
				Code:   http.StatusInternalServerError,
			})
			return
		}

		// PI stays "succeeded" after refund — mark canceled in our DB
		_ = h.paymentService.CancelPaymentInDB(sub.BillingID)

		canceledSub, err := h.subscriptionService.MarkCanceled(sub.BillingID)
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
			Data:   canceledSub,
			Writer: w,
			Reader: r,
			Code:   http.StatusOK,
		})
	}
}
