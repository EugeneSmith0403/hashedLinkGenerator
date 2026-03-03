package invoice

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
	InvoiceStatusVoid          InvoiceStatus = "void"
)

type Invoice struct {
	gorm.Model
	AccountID      uint  `json:"accountID" gorm:"index;not null"`
	SubscriptionID *uint `json:"subscriptionId" gorm:"index"`

	BillingID string `json:"billingId" gorm:"uniqueIndex;not null"`

	Status InvoiceStatus `json:"status" gorm:"type:varchar(20);not null"`

	AmountDue       int64  `json:"amountDue"`
	AmountPaid      int64  `json:"amountPaid"`
	AmountRemaining int64  `json:"amountRemaining"`
	Currency        string `json:"currency" gorm:"type:varchar(3);not null"`

	DueDate *time.Time `json:"dueDate"`
	PaidAt  *time.Time `json:"paidAt"`

	HostedInvoiceURL string `json:"hostedInvoiceUrl"`
	InvoicePDF       string `json:"invoicePdf"`

	ProviderMetadata datatypes.JSON `json:"providerMetadata"`
}
