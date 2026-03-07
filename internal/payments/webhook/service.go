package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"adv/go-http/internal/consts"
	"adv/go-http/internal/models"
	invoiceService "adv/go-http/internal/payments/invoice"
	stripeServices "adv/go-http/internal/payments/stripe/services"
	"adv/go-http/internal/payments/subscription"
	rabbitmq "adv/go-http/pkg/rabbitMq"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type WebhookServiceDeps struct {
	CustomerAccountService *stripeServices.CustomerAccountService
	InvoiceService         *invoiceService.InvoiceService
	SubscriptionService    *subscription.SubscriptionService
	RabbitMq               *rabbitmq.RabbitMq
}

type WebhookService struct {
	customerAccountService *stripeServices.CustomerAccountService
	invoiceService         *invoiceService.InvoiceService
	subscriptionService    *subscription.SubscriptionService
	rabbitMq               *rabbitmq.RabbitMq
}

func NewWebhookService(deps WebhookServiceDeps) *WebhookService {
	_, err := deps.RabbitMq.CreateExchange(&rabbitmq.ExchangeOptions{
		Name:    consts.PaymentIntentExchange,
		Type:    "direct",
		Durable: true,
	})

	if err != nil {
		fmt.Print(err)
	}

	_, _, errQ := deps.RabbitMq.CreateQueueWithBinding(&rabbitmq.QueueWithBindingOptions{
		QueueOptions: &rabbitmq.QueueOptions{
			Name:       consts.PaymentIntentQueueSucceed,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		Exchange:   consts.PaymentIntentExchange,
		RoutingKey: consts.PaymentIntentRouting,
	})

	if errQ != nil {
		fmt.Print(errQ)
	}

	return &WebhookService{
		customerAccountService: deps.CustomerAccountService,
		invoiceService:         deps.InvoiceService,
		subscriptionService:    deps.SubscriptionService,
		rabbitMq:               deps.RabbitMq,
	}
}

func (s *WebhookService) IsPaymentIntentEvent(t stripeGo.EventType) bool {
	switch t {
	case stripeGo.EventTypePaymentIntentCreated,
		stripeGo.EventTypePaymentIntentProcessing,
		stripeGo.EventTypePaymentIntentSucceeded,
		stripeGo.EventTypePaymentIntentPaymentFailed,
		stripeGo.EventTypePaymentIntentCanceled,
		stripeGo.EventTypePaymentIntentAmountCapturableUpdated,
		stripeGo.EventTypePaymentIntentRequiresAction,
		stripeGo.EventTypePaymentIntentPartiallyFunded:
		return true
	}
	return false
}

func (s *WebhookService) IsSetupIntentEvent(t stripeGo.EventType) bool {
	return t == stripeGo.EventTypeSetupIntentSucceeded
}

func (s *WebhookService) HandleSetupIntentEvent(event *stripeGo.Event) error {
	var si stripeGo.SetupIntent
	if err := json.Unmarshal(event.Data.Raw, &si); err != nil {
		return err
	}
	if si.Customer == nil || si.PaymentMethod == nil {
		return nil
	}
	log.Printf("[stripe] setup_intent.succeeded: customer=%s pm=%s\n", si.Customer.ID, si.PaymentMethod.ID)
	return s.customerAccountService.SetDefaultPaymentMethod(si.Customer.ID, si.PaymentMethod.ID)
}

func (s *WebhookService) IsSubscriptionEvent(t stripeGo.EventType) bool {
	switch t {
	case stripeGo.EventTypeCustomerSubscriptionCreated,
		stripeGo.EventTypeCustomerSubscriptionUpdated,
		stripeGo.EventTypeCustomerSubscriptionDeleted,
		stripeGo.EventTypeCustomerSubscriptionPaused,
		stripeGo.EventTypeCustomerSubscriptionResumed,
		stripeGo.EventTypeCustomerSubscriptionTrialWillEnd:
		return true
	}
	return false
}

func (s *WebhookService) HandlePaymentIntentEvent(event *stripeGo.Event) error {
	var pi stripeGo.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return err
	}

	log.Printf("[stripe] %s: id=%s amount=%d\n", event.Type, pi.ID, pi.Amount)

	body, err := json.Marshal(models.PaymentIntentMessage{
		EventType: models.PaymentIntentEventType(event.Type),
		Data:      pi,
	})
	if err != nil {
		return fmt.Errorf("marshal payment intent message: %w", err)
	}

	s.rabbitMq.Publish(&rabbitmq.PublisherOptions{
		Exchange:    consts.PaymentIntentExchange,
		RoutingKey:  consts.PaymentIntentRouting,
		Body:        body,
		ContentType: "application/json",
	})

	return nil
}

func (s *WebhookService) IsInvoiceEvent(t stripeGo.EventType) bool {
	return t == stripeGo.EventTypeInvoicePaymentSucceeded
}

func (s *WebhookService) HandleInvoiceEvent(event *stripeGo.Event) error {
	var inv stripeGo.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return err
	}

	switch event.Type {
	case stripeGo.EventTypeInvoicePaymentSucceeded:
		log.Printf("[stripe] invoice.payment_succeeded: id=%s amount=%d\n", inv.ID, inv.AmountPaid)

		var accountID uint
		var subscriptionID *uint
		if inv.Parent != nil &&
			inv.Parent.SubscriptionDetails != nil &&
			inv.Parent.SubscriptionDetails.Subscription != nil {
			billingID := inv.Parent.SubscriptionDetails.Subscription.ID
			sub, err := s.subscriptionService.GetByBillingID(billingID)
			if err != nil {
				return fmt.Errorf("get subscription: %w", err)
			}
			if sub != nil {
				accountID = sub.UserID
				subscriptionID = &sub.ID
			}
		}

		if err := s.invoiceService.CreatePaymentAndInvoiceFromStripeInvoice(&inv, accountID, subscriptionID); err != nil {
			log.Printf("[stripe] failed to create payment/invoice for invoice %s: %v\n", inv.ID, err)
		}
	}

	return nil
}

func (s *WebhookService) HandleSubscriptionEvent(event *stripeGo.Event) error {
	var sub stripeGo.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	switch event.Type {
	case stripeGo.EventTypeCustomerSubscriptionCreated:
		log.Printf("[stripe] subscription.created: id=%s status=%s\n", sub.ID, sub.Status)

		existing, err := s.subscriptionService.GetByBillingID(sub.ID)
		if err != nil {
			return err
		}

		var localSub *subscription.Subscription
		if existing == nil {
			userIDStr := sub.Metadata["user_id"]
			planIDStr := sub.Metadata["plan_id"]
			userID64, uErr := strconv.ParseUint(userIDStr, 10, 64)
			planID64, pErr := strconv.ParseUint(planIDStr, 10, 64)
			if uErr != nil || pErr != nil {
				log.Printf("[stripe] subscription.created: missing user_id/plan_id in metadata for %s\n", sub.ID)
				break
			}
			localSub, err = s.subscriptionService.CreateFromStripeSub(&sub, uint(userID64), uint(planID64))
			if err != nil {
				return fmt.Errorf("create subscription from event: %w", err)
			}
		} else {
			localSub = existing
		}

		if sub.LatestInvoice != nil && sub.LatestInvoice.ID != "" {
			if err := s.invoiceService.CreateInitialInvoice(sub.LatestInvoice.ID, localSub.UserID, localSub.ID); err != nil {
				log.Printf("[stripe] subscription.created: failed to create initial invoice for %s: %v\n", sub.ID, err)
			}
		}

	case stripeGo.EventTypeCustomerSubscriptionUpdated,
		stripeGo.EventTypeCustomerSubscriptionDeleted,
		stripeGo.EventTypeCustomerSubscriptionPaused,
		stripeGo.EventTypeCustomerSubscriptionResumed:
		log.Printf("[stripe] %s: id=%s status=%s\n", event.Type, sub.ID, sub.Status)
		if err := s.subscriptionService.UpdateSubscriptionFromEvent(&sub); err != nil {
			return err
		}

	case stripeGo.EventTypeCustomerSubscriptionTrialWillEnd:
		log.Printf("[stripe] subscription.trial_will_end: id=%s trial_end=%d\n", sub.ID, sub.TrialEnd)
	}

	return nil
}
