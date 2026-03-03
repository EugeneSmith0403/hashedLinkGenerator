package webhook

import (
	"io"
	"net/http"

	internalStripe "adv/go-http/internal/payments/stripe"
	"adv/go-http/internal/payments/subscription"
)

type WebhookHandlerDeps struct {
	StripeService       *internalStripe.StripeService
	SubscriptionService *subscription.SubscriptionService
}

type WebhookHandler struct {
	stripeService  *internalStripe.StripeService
	webhookService *WebhookService
}

func NewWebhookHandlers(router *http.ServeMux, deps WebhookHandlerDeps) {
	handler := &WebhookHandler{
		stripeService: deps.StripeService,
		webhookService: NewWebhookService(WebhookServiceDeps{
			StripeService:       deps.StripeService,
			SubscriptionService: deps.SubscriptionService,
		}),
	}

	router.HandleFunc("POST /stripe/webhook", handler.handleWebhook())
}

func (h *WebhookHandler) handleWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		sigHeader := r.Header.Get("Stripe-Signature")
		event, err := h.stripeService.DetectPaymentWebhook(payload, sigHeader)
		if err != nil {
			http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
			return
		}

		switch {
		case h.webhookService.IsPaymentIntentEvent(event.Type):
			if err := h.webhookService.HandlePaymentIntentEvent(event); err != nil {
				http.Error(w, "failed to process payment intent event", http.StatusInternalServerError)
				return
			}
		case h.webhookService.IsSubscriptionEvent(event.Type):
			if err := h.webhookService.HandleSubscriptionEvent(event); err != nil {
				http.Error(w, "failed to process subscription event", http.StatusInternalServerError)
				return
			}
		case h.webhookService.IsSetupIntentEvent(event.Type):
			if err := h.webhookService.HandleSetupIntentEvent(event); err != nil {
				http.Error(w, "failed to process setup intent event", http.StatusInternalServerError)
				return
			}
		case h.webhookService.IsInvoiceEvent(event.Type):
			if err := h.webhookService.HandleInvoiceEvent(event); err != nil {
				http.Error(w, "failed to process invoice event", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
