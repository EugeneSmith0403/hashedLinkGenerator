package payments

import (
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Payment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`

	AccountID      uint  `json:"accountId" gorm:"index;not null"`
	InvoiceID      *uint `json:"invoiceId" gorm:"index"`
	PlanID         *uint `json:"planId" gorm:"index"`
	SubscriptionID *uint `json:"subscriptionId" gorm:"index"`

	PaymentIntentID string  `json:"paymentIntentId" gorm:"uniqueIndex"`
	ChargeID        *string `json:"chargeId" gorm:"uniqueIndex"`

	Amount      int64  `json:"amount"`
	PlatformFee int64  `json:"platformFee"`
	NetAmount   int64  `json:"netAmount"`
	Currency    string `json:"currency" gorm:"type:varchar(3);not null"`

	Status            stripe.PaymentIntentStatus `json:"status" gorm:"type:varchar(40);not null"`
	PaymentMethodType string                     `json:"paymentMethodType"`

	FailureCode    string `json:"failureCode"`
	FailureMessage string `json:"failureMessage"`

	ProviderMetadata datatypes.JSON `json:"providerMetadata"`
}

type PaymentUpdate struct {
	Status           stripe.PaymentIntentStatus `json:"status"`
	ChargeID         *string                    `json:"chargeId"`
	FailureCode      string                     `json:"failureCode"`
	FailureMessage   string                     `json:"failureMessage"`
	ProviderMetadata datatypes.JSON             `json:"providerMetadata"`
}
