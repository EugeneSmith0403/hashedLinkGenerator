package auth

import (
	"errors"
	"fmt"
	"link-generator/configs"
	internalJWT "link-generator/internal/jwt"
	"link-generator/internal/models"
	"link-generator/pkg/helpers"
	"link-generator/pkg/redis"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type IUserRepository interface {
	Create(user *models.User) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
}

type AuthServiceDeps struct {
	UserRepository IUserRepository
	Config         *configs.Config
	JWTService     *internalJWT.JWTService
	RedisService   *redis.Redis
}

type AuthService struct {
	UserRepository IUserRepository
	Config         *configs.Config
	JWTService     *internalJWT.JWTService
	RedisService   *redis.Redis
}

func NewAuthService(deps AuthServiceDeps) *AuthService {
	return &AuthService{
		UserRepository: deps.UserRepository,
		Config:         deps.Config,
		JWTService:     deps.JWTService,
		RedisService:   deps.RedisService,
	}
}

func (service *AuthService) GenerateToken(email string) (string, time.Time, error) {
	now := time.Now()
	expiredHours, err := strconv.Atoi(service.Config.Auth.ExpiredAt)
	if err != nil {
		return "", time.Time{}, err
	}

	expiryTime := now.Add(helpers.ToHours(expiredHours))

	claims := jwt.MapClaims{
		"email":     email,
		"createdAt": now,
		"exp":       expiryTime.Unix(),
		"iat":       now.Unix(),
	}

	token, tokenErr := service.JWTService.GenerateToken(&claims)
	if tokenErr != nil {
		return "", time.Time{}, tokenErr
	}

	if service.RedisService != nil {
		userKey := fmt.Sprintf("token:%s", email)
		service.RedisService.Set(userKey, true, helpers.ToHours(expiredHours))
	}

	return token, expiryTime, nil
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

	userModel := &models.User{Name: name, Email: email, Password: string(cryptedPassword)}

	createdUser, err := service.UserRepository.Create(userModel)

	if err != nil {
		return "", err
	}

	return createdUser.Email, nil
}
