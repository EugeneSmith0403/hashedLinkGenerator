package subscription

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive            SubscriptionStatus = "active"
	SubscriptionStatusCanceled          SubscriptionStatus = "canceled"
	SubscriptionStatusPastDue           SubscriptionStatus = "past_due"
	SubscriptionStatusTrialing          SubscriptionStatus = "trialing"
	SubscriptionStatusUnpaid            SubscriptionStatus = "unpaid"
	SubscriptionStatusIncomplete        SubscriptionStatus = "incomplete"
	SubscriptionStatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
	SubscriptionStatusPaused            SubscriptionStatus = "paused"
)

type Subscription struct {
	gorm.Model
	UserID uint `json:"userId" gorm:"index;not null"`
	PlanID uint `json:"planId" gorm:"index;not null"`

	BillingID  string `json:"billingId" gorm:"uniqueIndex;not null"`
	CustomerID string `json:"customerId" gorm:"index;not null"`

	Status             SubscriptionStatus `json:"status" gorm:"type:varchar(30);not null"`
	CurrentPeriodStart time.Time          `json:"currentPeriodStart"`
	CurrentPeriodEnd   time.Time          `json:"currentPeriodEnd"`

	CancelAt   *time.Time `json:"cancelAt"`
	CanceledAt *time.Time `json:"canceledAt"`

	TrialStart *time.Time `json:"trialStart"`
	TrialEnd   *time.Time `json:"trialEnd"`

	ProviderMetadata datatypes.JSON `json:"providerMetadata"`
}
