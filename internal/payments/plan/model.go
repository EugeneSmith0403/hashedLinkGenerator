package plan

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Plan struct {
	gorm.Model
	IsActive        bool           `json:"isActive"`
	Name            string         `json:"name"`
	Cost            float32        `json:"cost"`
	Currency        string         `json:"currency"`
	IntervalMonths  int            `json:"intervalMonths" gorm:"not null;default:1"`
	Features        datatypes.JSON `json:"features"`
	StripePriceID   string         `json:"stripePriceId"`
	StripeProductID string         `json:"stripeProductId"`
}
