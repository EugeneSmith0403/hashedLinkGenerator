package di

import "link-generator/internal/user"

type IUserRepository interface {
	Create(u *user.User) (*user.User, error)
	FindByEmail(email string) (*user.User, error)
}
