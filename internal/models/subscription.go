package models

import stripeGo "github.com/stripe/stripe-go/v84"

type SubscriptionEventType string

const (
	SubscriptionCreated               SubscriptionEventType = "customer.subscription.created"
	SubscriptionUpdated               SubscriptionEventType = "customer.subscription.updated"
	SubscriptionDeleted               SubscriptionEventType = "customer.subscription.deleted"
	SubscriptionPaused                SubscriptionEventType = "customer.subscription.paused"
	SubscriptionResumed               SubscriptionEventType = "customer.subscription.resumed"
	SubscriptionTrialWillEnd          SubscriptionEventType = "customer.subscription.trial_will_end"
	SubscriptionPendingUpdateApplied  SubscriptionEventType = "customer.subscription.pending_update_applied"
	SubscriptionPendingUpdateExpired  SubscriptionEventType = "customer.subscription.pending_update_expired"
)

type SubscriptionMessage struct {
	EventType SubscriptionEventType    `json:"event_type"`
	Data      stripeGo.Subscription    `json:"data"`
}
