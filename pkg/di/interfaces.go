package di

import "link-generator/internal/models"

type IUserRepository interface {
	Create(u *models.User) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
}
