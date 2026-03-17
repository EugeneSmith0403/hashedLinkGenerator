package authsession

import (
	"link-generator/internal/models"
	"link-generator/pkg/db"
)

type AuthSessionRepository struct {
	db *db.Db
}

func NewAuthSessionRepository(db *db.Db) *AuthSessionRepository {
	return &AuthSessionRepository{db}
}

func (a *AuthSessionRepository) Insert(authSession *models.AuthSession) (*models.AuthSession, error) {
	result := a.db.DB.Save(&authSession)

	if result.Error != nil {
		return nil, result.Error
	}

	return authSession, nil
}

func (a *AuthSessionRepository) UpdateByAccountId(authSessionData *models.AuthSession) (*models.AuthSession, error) {

	var authSessionResult *models.AuthSession

	res := a.db.DB.Model(&models.AuthSession{}).
		Select("is_active", "is_verify").
		Where("account_id=?", authSessionData.AccountID).
		Updates(authSessionData)

	if res.Error != nil {
		return nil, res.Error
	}

	return authSessionResult, nil
}

func (a *AuthSessionRepository) GetByToken(token string) (*models.AuthSessionWithRelations, error) {
	var result *models.AuthSessionWithRelations

	res := a.db.DB.
		Preload("Account").
		Preload("Account.User").
		First(&result, "token=? AND is_active=true", token)

	if res.Error != nil {
		return nil, res.Error
	}

	return result, nil
}
