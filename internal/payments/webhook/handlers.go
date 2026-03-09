package webhook

import (
	"io"
	"net/http"

	"adv/go-http/internal/account"
	invoiceService "adv/go-http/internal/payments/invoice"
	stripeServices "adv/go-http/internal/payments/stripe/services"
	"adv/go-http/internal/payments/subscription"
	rabbitmq "adv/go-http/pkg/rabbitMq"
)

type WebhookHandlerDeps struct {
	PaymentService         *stripeServices.PaymentService
	CustomerAccountService *stripeServices.CustomerAccountService
	InvoiceService         *invoiceService.InvoiceService
	SubscriptionService    *subscription.SubscriptionService
	AccountRepository      *account.AccountRepository
	RabbitMq               *rabbitmq.RabbitMq
}

type WebhookHandler struct {
	paymentService *stripeServices.PaymentService
	webhookService *WebhookService
}

func NewWebhookHandlers(router *http.ServeMux, deps WebhookHandlerDeps) {
	handler := &WebhookHandler{
		paymentService: deps.PaymentService,
		webhookService: NewWebhookService(WebhookServiceDeps{
			CustomerAccountService: deps.CustomerAccountService,
			InvoiceService:         deps.InvoiceService,
			SubscriptionService:    deps.SubscriptionService,
			AccountRepository:      deps.AccountRepository,
			RabbitMq:               deps.RabbitMq,
		}),
	}

	router.Handle("POST /stripe/webhook", handler.handleWebhook())
}

func (h *WebhookHandler) handleWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		sigHeader := r.Header.Get("Stripe-Signature")
		event, err := h.paymentService.DetectPaymentWebhook(payload, sigHeader)
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
