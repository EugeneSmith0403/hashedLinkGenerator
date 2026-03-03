package account

import (
	"adv/go-http/pkg/db"
	"errors"

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

func (r AccountRepository) Create(account *Account) (*Account, error) {

	result := r.db.DB.Create(account)

	if result.Error != nil {
		return nil, result.Error
	}

	return account, nil

}

func (r *AccountRepository) FindById(id uint) (*Account, error) {
	var account Account
	result := r.db.DB.First(&account, "id=?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &account, nil

}

func (r *AccountRepository) FindByUserId(userId uint) (*Account, error) {
	var account Account
	result := r.db.DB.First(&account, "user_id=?", userId)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &account, nil
}

func (r *AccountRepository) Update(account *Account) (*Account, error) {
	result := r.db.DB.Save(account)
	if result.Error != nil {
		return nil, result.Error
	}
	return account, nil
}
