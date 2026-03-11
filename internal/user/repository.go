package user

import (
	"errors"
	"link-generator/internal/models"
	"link-generator/pkg/db"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *db.Db
}

func NewUserRepository(db *db.Db) *UserRepository {
	return &UserRepository{
		db,
	}
}

func (r UserRepository) Create(user *models.User) (*models.User, error) {

	result := r.db.DB.Create(user)

	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil

}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.DB.First(&user, "email=?", email)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil

}

func (r *UserRepository) FindByEmailWithAccounts(email string) (*models.User, error) {
	var user models.User
	result := r.db.DB.Preload("Accounts").First(&user, "email=?", email)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil

}
