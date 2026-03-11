package user

import "link-generator/internal/models"

func NewUser(name string, email string, password string) *models.User {
	return &models.User{
		Name:     name,
		Email:    email,
		Password: password,
	}
}
