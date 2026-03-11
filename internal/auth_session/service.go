package authsession

import (
	"link-generator/internal/models"
	"link-generator/pkg/db"
	"time"
)

type AuthSessionServiceDeps struct{}

type AuthSessionService struct {
	AuthSessionRepository AuthSessionRepository
}

type AddOptions struct {
	AccountID uint
	Token     string
	IpAddress string
	UserAgent string
	IsVerify  bool
	ExpiresAt time.Time
}

func NewAuthSessionService(db *db.Db) *AuthSessionService {
	return &AuthSessionService{
		AuthSessionRepository: *NewAuthSessionRepository(db),
	}
}

func (a AuthSessionService) Update(options *AddOptions) (*models.AuthSession, error) {

	a.AuthSessionRepository.UpdateByAccountId(&models.AuthSession{AccountID: options.AccountID, Is_active: false, Is_verify: false})

	session := &models.AuthSession{
		AccountID:  options.AccountID,
		Is_active:  true,
		Token:      options.Token,
		Ip_address: options.IpAddress,
		User_agent: options.UserAgent,
		Is_verify:  options.IsVerify,
		Expires_at: options.ExpiresAt,
	}

	res, err := a.AuthSessionRepository.Insert(session)

	if err != nil {
		return nil, err
	}

	return res, nil
}
