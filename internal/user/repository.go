package user

import (
	"adv/go-http/pkg/db"
	"errors"

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

func (r UserRepository) Create(user *User) (*User, error) {

	result := r.db.DB.Create(user)

	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil

}

func (r *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	result := r.db.DB.First(&user, "email=?", email)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil

}
