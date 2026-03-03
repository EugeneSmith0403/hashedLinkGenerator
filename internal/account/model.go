package account

import (
	"time"

	"gorm.io/gorm"
)

type AccountStatus string

const (
	StatusActive   AccountStatus = "active"
	StatusInactive AccountStatus = "inactive"
	StatusPending  AccountStatus = "pending"
	StatusBanned   AccountStatus = "banned"
)

type PaymentProvider string

const (
	ProviderStripe   PaymentProvider = "stripe"
	ProviderPayPal   PaymentProvider = "paypal"
	ProviderYooKassa PaymentProvider = "yookassa"
)

type Account struct {
	gorm.Model
	UserID        uint            `json:"userId" gorm:"uniqueIndex:idx_account_userId"`
	AccountStatus AccountStatus   `json:"accountStatus" gorm:"type:varchar(20);not null"`
	Provider      PaymentProvider `json:"provider" gorm:"type:varchar(20);not null"`
	CustomerID    string          `json:"customerId"`
	BannedBy      string          `json:"bannedBy"`
	BannedAt      *time.Time      `json:"bannedAt"`
}
