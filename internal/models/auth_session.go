package models

import (
	"time"

	"gorm.io/gorm"
)

type AuthSession struct {
	gorm.Model
	AccountID  uint          `json:"accountId"`
	Token      string        `json:"token"`
	Ip_address string        `json:"ipAddress"`
	User_agent string        `json:"userAgent"`
	Is_active  bool          `json:"isActive"`
	Is_verify  bool          `json:"isVerify"`
	Expires_at time.Time `json:"expiresAt"`
}

type AuthSessionWithRelations struct {
	AuthSession
	Account Account `json:"account"`
}

func (AuthSessionWithRelations) TableName() string {
	return "auth_sessions"
}
