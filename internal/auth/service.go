package auth

import (
	"errors"
	"fmt"
	"link-generator/configs"
	internalJWT "link-generator/internal/jwt"
	"link-generator/internal/user"
	"link-generator/pkg/helpers"
	"link-generator/pkg/redis"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceDeps struct {
	UserRepository *user.UserRepository
	Config         *configs.Config
	JWTService     *internalJWT.JWTService
	RedisService   *redis.Redis
}

type AuthService struct {
	UserRepository *user.UserRepository
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

func (service *AuthService) GenerateToken(email string) (string, error) {
	now := time.Now()
	expiredHours, err := strconv.Atoi(service.Config.Auth.ExpiredAt)
	if err != nil {
		return "", err
	}

	expirationTime := now.Add(helpers.ToHours(expiredHours)).Unix()

	claims := jwt.MapClaims{
		"email":     email,
		"createdAt": now,
		"exp":       expirationTime,
		"iat":       now.Unix(),
	}

	token, tokenErr := service.JWTService.GenerateToken(&claims)
	if tokenErr != nil {
		return "", tokenErr
	}

	userKey := fmt.Sprintf("token:%s", email)
	service.RedisService.Set(userKey, true, time.Duration(expirationTime))

	return token, nil
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
