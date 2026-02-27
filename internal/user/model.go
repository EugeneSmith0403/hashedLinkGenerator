package user

import "gorm.io/gorm"

type User struct {
	gorm.Model `gorm:"index"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	Name       string `json:"name" validate:"required"`
}

func NewUser(name string, email string, password string) *User {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
	}
}
