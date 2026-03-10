package webhook

import (
	"encoding/json"
	"log"

	"link-generator/internal/account"
	"link-generator/internal/models"
	invoiceService "link-generator/internal/payments/invoice"
	stripeServices "link-generator/internal/payments/stripe/services"
	"link-generator/internal/payments/subscription"
	"link-generator/internal/publishers"
	rabbitmq "link-generator/pkg/rabbitMq"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type WebhookServiceDeps struct {
	CustomerAccountService *stripeServices.CustomerAccountService
	InvoiceService         *invoiceService.InvoiceService
	SubscriptionService    *subscription.SubscriptionService
	AccountRepository      *account.AccountRepository
	RabbitMq               *rabbitmq.RabbitMq
}

type WebhookService struct {
	customerAccountService *stripeServices.CustomerAccountService
	invoiceService         *invoiceService.InvoiceService
	subscriptionService    *subscription.SubscriptionService
	accountRepository      *account.AccountRepository
	rabbitMq               *rabbitmq.RabbitMq
	paymentPublisher       *publishers.PaymentPublisher
	subscriptionPublisher  *publishers.SubscriptionPublisher
	invoicePublisher       *publishers.InvoicePublisher
}

func NewWebhookService(deps WebhookServiceDeps) *WebhookService {
	paymentPub := publishers.NewPaymentPublisher(deps.RabbitMq)
	subscriptionPub := publishers.NewSubscriptionPublisher(deps.RabbitMq)
	invoicePub := publishers.NewInvoicePublisher(deps.RabbitMq)

	paymentPub.CreateExchangeAndQueue()
	subscriptionPub.CreateExchangeAndQueue()
	invoicePub.CreateExchangeAndQueue()

	return &WebhookService{
		customerAccountService: deps.CustomerAccountService,
		invoiceService:         deps.InvoiceService,
		subscriptionService:    deps.SubscriptionService,
		accountRepository:      deps.AccountRepository,
		rabbitMq:               deps.RabbitMq,
		paymentPublisher:       paymentPub,
		subscriptionPublisher:  subscriptionPub,
		invoicePublisher:       invoicePub,
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
	err := s.paymentPublisher.PublishDataRawToQueue(models.PaymentIntentEventType(event.Type), event.Data)

	if err != nil {
		return err
	}

	return nil
}

func (s *WebhookService) HandleSubscriptionEvent(event *stripeGo.Event) error {

	err := s.subscriptionPublisher.PublishDataRawToQueue(models.SubscriptionEventType(event.Type), event.Data)

	if err != nil {
		return err
	}

	return nil
}

func (s *WebhookService) IsInvoiceEvent(t stripeGo.EventType) bool {
	return t == stripeGo.EventTypeInvoicePaymentSucceeded
}

func (s *WebhookService) HandleInvoiceEvent(event *stripeGo.Event) error {
	return s.invoicePublisher.PublishDataRawToQueue(models.InvoiceEventType(event.Type), event.Data)
}
