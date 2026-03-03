package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	internalStripe "adv/go-http/internal/payments/stripe"
	"adv/go-http/internal/payments/subscription"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type WebhookServiceDeps struct {
	StripeService       *internalStripe.StripeService
	SubscriptionService *subscription.SubscriptionService
}

type WebhookService struct {
	stripeService       *internalStripe.StripeService
	subscriptionService *subscription.SubscriptionService
}

func NewWebhookService(deps WebhookServiceDeps) *WebhookService {
	return &WebhookService{
		stripeService:       deps.StripeService,
		subscriptionService: deps.SubscriptionService,
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
	return s.stripeService.SetDefaultPaymentMethod(si.Customer.ID, si.PaymentMethod.ID)
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

	switch event.Type {
	case stripeGo.EventTypePaymentIntentCreated:
		log.Printf("[stripe] payment_intent.created: id=%s amount=%d currency=%s\n", pi.ID, pi.Amount, pi.Currency)

	case stripeGo.EventTypePaymentIntentProcessing:
		log.Printf("[stripe] payment_intent.processing: id=%s\n", pi.ID)

	case stripeGo.EventTypePaymentIntentSucceeded:
		log.Printf("[stripe] payment_intent.succeeded: id=%s amount=%d\n", pi.ID, pi.Amount)

		if _, ok := pi.Metadata["payment_id"]; !ok {
			// subscription-generated PI — handled by invoice.payment_succeeded
			break
		}
		updatedPi, err := s.stripeService.UpdatePaymentFromIntent(&pi)
		if err != nil {
			log.Printf("[stripe] failed to update payment for payment_intent %s: %v\n", pi.ID, err)
			break
		}
		// TODO send to invoice queue
		if err := s.stripeService.CreateInvoiceForPayment(updatedPi); err != nil {
			log.Printf("[stripe] failed to create invoice for payment_intent %s: %v\n", pi.ID, err)
		}

	case stripeGo.EventTypePaymentIntentPaymentFailed:
		var reason string
		if pi.LastPaymentError != nil {
			reason = pi.LastPaymentError.Msg
		}
		log.Printf("[stripe] payment_intent.payment_failed: id=%s reason=%s\n", pi.ID, reason)
		if _, err := s.stripeService.UpdatePaymentFromIntent(&pi); err != nil {
			return err
		}

	case stripeGo.EventTypePaymentIntentCanceled:
		log.Printf("[stripe] payment_intent.canceled: id=%s\n", pi.ID)

	case stripeGo.EventTypePaymentIntentAmountCapturableUpdated:
		log.Printf("[stripe] payment_intent.amount_capturable_updated: id=%s capturable=%d\n", pi.ID, pi.AmountCapturable)

	case stripeGo.EventTypePaymentIntentRequiresAction:
		log.Printf("[stripe] payment_intent.requires_action: id=%s\n", pi.ID)

	case stripeGo.EventTypePaymentIntentPartiallyFunded:
		log.Printf("[stripe] payment_intent.partially_funded: id=%s received=%d\n", pi.ID, pi.AmountReceived)
	}

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

		if err := s.stripeService.CreatePaymentAndInvoiceFromStripeInvoice(&inv, accountID, subscriptionID); err != nil {
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
			if err := s.stripeService.CreateInitialInvoice(sub.LatestInvoice.ID, localSub.UserID, localSub.ID); err != nil {
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
