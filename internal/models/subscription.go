package models

import (
	"time"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type SubscriptionEventType string

const (
	SubscriptionCreated              SubscriptionEventType = "customer.subscription.created"
	SubscriptionUpdated              SubscriptionEventType = "customer.subscription.updated"
	SubscriptionDeleted              SubscriptionEventType = "customer.subscription.deleted"
	SubscriptionPaused               SubscriptionEventType = "customer.subscription.paused"
	SubscriptionResumed              SubscriptionEventType = "customer.subscription.resumed"
	SubscriptionTrialWillEnd         SubscriptionEventType = "customer.subscription.trial_will_end"
	SubscriptionPendingUpdateApplied SubscriptionEventType = "customer.subscription.pending_update_applied"
	SubscriptionPendingUpdateExpired SubscriptionEventType = "customer.subscription.pending_update_expired"
)

type SubscriptionMessage struct {
	EventType SubscriptionEventType `json:"event_type"`
	Data      stripeGo.Subscription `json:"data"`
}

type SubscriptionInfo struct {
	ID                 uint       `json:"id"`
	CreatedAt          time.Time  `json:"createdAt"`
	PlanID             uint       `json:"planId"`
	Status             string     `json:"status"`
	CurrentPeriodStart time.Time  `json:"currentPeriodStart"`
	CurrentPeriodEnd   time.Time  `json:"currentPeriodEnd"`
	CancelAt           *time.Time `json:"cancelAt"`
	CanceledAt         *time.Time `json:"canceledAt"`
	TrialStart         *time.Time `json:"trialStart"`
	TrialEnd           *time.Time `json:"trialEnd"`
	IsPaymentIntent    bool       `json:"isPaymentIntent"`
}

type ISubscriptionService interface {
	GetSubscriptionByUserID(userID uint) (*SubscriptionInfo, error)
}
