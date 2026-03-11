package account

import (
	"errors"
	"link-generator/internal/models"
	"link-generator/pkg/db"

	"gorm.io/gorm"
)

type AccountRepository struct {
	db *db.Db
}

func NewAccountRepository(db *db.Db) *AccountRepository {
	return &AccountRepository{
		db,
	}
}

func (r AccountRepository) Create(account *models.Account) (*models.Account, error) {

	result := r.db.DB.Create(account)

	if result.Error != nil {
		return nil, result.Error
	}

	return account, nil

}

func (r *AccountRepository) FindById(id uint) (*models.Account, error) {
	var account models.Account
	result := r.db.DB.First(&account, "id=?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &account, nil

}

func (r *AccountRepository) FindByUserId(userId uint) (*models.Account, error) {
	var account models.Account
	result := r.db.DB.First(&account, "user_id=?", userId)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &account, nil
}

func (r *AccountRepository) Update(account *models.Account) (*models.Account, error) {
	result := r.db.DB.Save(account)
	if result.Error != nil {
		return nil, result.Error
	}
	return account, nil
}
