package auth

import (
	"link-generator/internal/user"
	"link-generator/pkg/di"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository di.IUserRepository
}

func NewAuthService(userRep di.IUserRepository) *AuthService {
	return &AuthService{
		UserRepository: userRep,
	}
}

func (service *AuthService) Login(email, password string) bool {
	existedUser, err := service.UserRepository.FindByEmail(email)

	if err != nil || existedUser == nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(existedUser.Password), []byte(password)) == nil
}

func (service *AuthService) Register(name, email, password string) (string, error) {

	existedUser, err := service.UserRepository.FindByEmail(email)

	if err != nil {
		return "", err
	}

	if existedUser != nil {
		return "", errors.New(UserExists)
	}

	// encrypte password
	cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	userModel := user.NewUser(name, email, string(cryptedPassword))

	createdUser, err := service.UserRepository.Create(userModel)

	if err != nil {
		return "", err
	}

	return createdUser.Email, nil
}
