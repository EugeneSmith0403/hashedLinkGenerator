package models

import (
	"time"

	"gorm.io/gorm"
)

type AccountInfo struct {
	AccountID    uint
	UserID       uint
	Is2FAEnabled bool
	TotpSecret   string
}

type IAccountService interface {
	CreateAccount(userId uint, name, email string) (*Account, error)
	GetAccountInfoByEmail(email string) (*AccountInfo, error)
	Setup2FA(email string) (string, error)
	Verify2Fa(code, email string) bool
}

type IAuthService interface {
	GenerateToken(email string) (string, time.Time, error)
}

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
	TotpSecret    string          `json:"totpSecret" gorm:"column:totp_secret"`
	Is2FAEnabled  bool            `json:"is2faEnabled" gorm:"column:is_2fa_enabled;default:false"`
	AuthSession   []AuthSession   `json:"authSession"`
	User          User            `json:"user"`
}
