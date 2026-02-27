package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JwtData struct {
	Email string
}

type JwtDeps struct {
	Secret string
}

type JWTService struct {
	*JwtDeps
}

func NewJWTService(deps JwtDeps) *JWTService {
	return &JWTService{
		JwtDeps: &JwtDeps{
			Secret: deps.Secret,
		},
	}
}

func (service JWTService) GenerateToken(claims *jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(service.Secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (service *JWTService) CheckToken(token string, claims *jwt.MapClaims) (*jwt.MapClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(service.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	parsedClaims, ok := parsedToken.Claims.(*jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return parsedClaims, nil
}

func (service *JWTService) ParseEmail(token string) (bool, *JwtData) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(service.Secret), nil
	})

	if err != nil {
		return false, nil
	}

	email := t.Claims.(jwt.MapClaims)["email"]

	jwtData := &JwtData{
		Email: email.(string),
	}

	return t.Valid, jwtData

}
