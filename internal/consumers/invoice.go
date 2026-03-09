package consumers

import (
	"encoding/json"
	"fmt"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"

	"adv/go-http/internal/account"
	"adv/go-http/internal/models"
	"adv/go-http/internal/payments/invoice"
	"adv/go-http/internal/payments/subscription"
)

type InvoiceConsumerDeps struct {
	InvoiceSvc        *invoice.InvoiceService
	SubscriptionSvc   *subscription.SubscriptionService
	AccountRepository *account.AccountRepository
}

type InvoiceConsumer struct {
	invoiceSvc        *invoice.InvoiceService
	subscriptionSvc   *subscription.SubscriptionService
	accountRepository *account.AccountRepository
}

func NewInvoiceConsumer(deps *InvoiceConsumerDeps) *InvoiceConsumer {
	return &InvoiceConsumer{
		invoiceSvc:        deps.InvoiceSvc,
		subscriptionSvc:   deps.SubscriptionSvc,
		accountRepository: deps.AccountRepository,
	}
}

func (c *InvoiceConsumer) Handle(body []byte) error {
	var msg models.InvoiceMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	switch msg.EventType {
	case models.InvoicePaymentSucceeded:
		return c.handlePaymentSucceeded(&msg.Data)
	default:
		log.Printf("[consumer] no handler for %s, skipping %s", msg.EventType, msg.Data.ID)
		return nil
	}
}

func (c *InvoiceConsumer) handlePaymentSucceeded(inv *stripeGo.Invoice) error {
	log.Printf("[stripe] invoice.payment_succeeded: id=%s amount=%d\n", inv.ID, inv.AmountPaid)

	var accountID uint
	var subscriptionID *uint
	if inv.Parent != nil &&
		inv.Parent.SubscriptionDetails != nil &&
		inv.Parent.SubscriptionDetails.Subscription != nil {
		billingID := inv.Parent.SubscriptionDetails.Subscription.ID
		sub, err := c.subscriptionSvc.GetByBillingID(billingID)
		if err != nil {
			return fmt.Errorf("get subscription: %w", err)
		}
		if sub != nil {
			subscriptionID = &sub.ID
			if acct, aErr := c.accountRepository.FindByUserId(sub.UserID); aErr == nil && acct != nil {
				accountID = acct.ID
			}
		}
	}

	return c.invoiceSvc.CreatePaymentAndInvoiceFromStripeInvoice(inv, accountID, subscriptionID)
}
