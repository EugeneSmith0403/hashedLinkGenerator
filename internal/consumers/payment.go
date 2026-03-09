package consumers

import (
	"encoding/json"
	"fmt"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"

	"adv/go-http/internal/models"
	"adv/go-http/internal/payments/invoice"
	stripeServices "adv/go-http/internal/payments/stripe/services"
	"adv/go-http/internal/payments/subscription"
)

type PaymentIntentConsumerDeps struct {
	PaymentSvc      *stripeServices.PaymentService
	InvoiceSvc      *invoice.InvoiceService
	SubscriptionSvc *subscription.SubscriptionService
}

type PaymentIntentConsumer struct {
	paymentSvc      *stripeServices.PaymentService
	invoiceSvc      *invoice.InvoiceService
	subscriptionSvc *subscription.SubscriptionService
}

func NewPaymentIntentConsumer(deps PaymentIntentConsumerDeps) *PaymentIntentConsumer {
	return &PaymentIntentConsumer{
		paymentSvc:      deps.PaymentSvc,
		invoiceSvc:      deps.InvoiceSvc,
		subscriptionSvc: deps.SubscriptionSvc,
	}
}

func (c *PaymentIntentConsumer) Handle(body []byte) error {
	var msg models.PaymentIntentMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	switch msg.EventType {
	case models.PaymentIntentSucceeded:
		return c.handleSucceeded(&msg.Data)
	case models.PaymentIntentPaymentFailed:
		return c.handleFailed(&msg.Data)
	default:
		log.Printf("[consumer] no handler for %s, skipping %s", msg.EventType, msg.Data.ID)
		return nil
	}
}

func (c *PaymentIntentConsumer) handleSucceeded(pi *stripeGo.PaymentIntent) error {
	if _, ok := pi.Metadata["payment_id"]; !ok {
		// subscription-generated PI — handled by invoice.payment_succeeded
		return nil
	}

	updatedPi, err := c.paymentSvc.UpdatePaymentFromIntent(pi)
	if err != nil {
		return fmt.Errorf("update payment: %w", err)
	}

	sub, err := c.subscriptionSvc.CreateFromPaymentIntent(pi)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	log.Printf("[consumer] subscription created/found: id=%d", sub.ID)

	if err := c.invoiceSvc.CreateInvoiceForPayment(updatedPi); err != nil {
		return fmt.Errorf("create invoice: %w", err)
	}

	return nil
}

func (c *PaymentIntentConsumer) handleFailed(pi *stripeGo.PaymentIntent) error {
	if _, err := c.paymentSvc.UpdatePaymentFromIntent(pi); err != nil {
		return fmt.Errorf("update payment: %w", err)
	}
	return nil
}
