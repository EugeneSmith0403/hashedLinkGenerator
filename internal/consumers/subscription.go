package consumers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	stripeGo "github.com/stripe/stripe-go/v84"

	"link-generator/internal/account"
	"link-generator/internal/models"
	"link-generator/internal/payments/invoice"
	"link-generator/internal/payments/subscription"
)

type SubscriptionConsumerDeps struct {
	SubscriptionSvc   *subscription.SubscriptionService
	InvoiceSvc        *invoice.InvoiceService
	AccountRepository *account.AccountRepository
}

type SubscriptionConsumer struct {
	subscriptionSvc   *subscription.SubscriptionService
	invoiceSvc        *invoice.InvoiceService
	accountRepository *account.AccountRepository
}

func NewSubscriptionConsumer(deps *SubscriptionConsumerDeps) *SubscriptionConsumer {
	return &SubscriptionConsumer{
		subscriptionSvc:   deps.SubscriptionSvc,
		invoiceSvc:        deps.InvoiceSvc,
		accountRepository: deps.AccountRepository,
	}
}

func (c *SubscriptionConsumer) Handle(body []byte) error {
	var msg models.SubscriptionMessage

	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	switch msg.EventType {
	case models.SubscriptionCreated:
		return c.handleSucceeded(&msg.Data)
	case models.SubscriptionUpdated:
	case models.SubscriptionPaused:
	case models.SubscriptionResumed:
		return c.handleUpdated(&msg.Data)
	case models.SubscriptionDeleted:
		return c.handleDeleted(&msg.Data)
	default:
		log.Printf("[consumer] no handler for %s, skipping %s", msg.EventType, msg.Data.ID)
		return nil
	}

	return nil
}

func (c *SubscriptionConsumer) handleSucceeded(sub *stripeGo.Subscription) error {

	log.Printf("[stripe] subscription.created: id=%s status=%s\n", sub.ID, sub.Status)

	existing, err := c.subscriptionSvc.GetByBillingID(sub.ID)
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
		}
		localSub, err = c.subscriptionSvc.CreateFromStripeSub(sub, uint(userID64), uint(planID64))
		if err != nil {
			return fmt.Errorf("create subscription from event: %w", err)
		}
	} else {
		localSub = existing
	}

	if sub.LatestInvoice != nil && sub.LatestInvoice.ID != "" {
		var accountID uint
		if c.accountRepository != nil {
			if acct, aErr := c.accountRepository.FindByUserId(localSub.UserID); aErr == nil && acct != nil {
				accountID = acct.ID
			}
		}
		if err := c.invoiceSvc.CreateInitialInvoice(sub.LatestInvoice.ID, accountID, localSub.ID); err != nil {
			log.Printf("[stripe] subscription.created: failed to create initial invoice for %s: %v\n", sub.ID, err)
		}
	}

	return nil
}

func (c *SubscriptionConsumer) handleUpdated(sub *stripeGo.Subscription) error {
	if err := c.subscriptionSvc.UpdateSubscriptionFromEvent(sub); err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	return nil
}

func (c *SubscriptionConsumer) handleDeleted(sub *stripeGo.Subscription) error {
	if _, err := c.subscriptionSvc.MarkCanceled(sub.ID); err != nil {
		return fmt.Errorf("mark canceled: %w", err)
	}

	if err := c.subscriptionSvc.UpdateSubscriptionFromEvent(sub); err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	return nil
}
