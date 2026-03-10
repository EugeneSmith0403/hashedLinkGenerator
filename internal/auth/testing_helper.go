package auth

import (
	"link-generator/configs"
	internalJWT "link-generator/internal/jwt"
	"link-generator/pkg/response"
)

// NewAuthHandlerForTest создает AuthHandler для тестов
func NewAuthHandlerForTest(config *configs.Config, authService *AuthService, jwtService *internalJWT.JWTService) *AuthHandler {
	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}

	return &AuthHandler{
		Config:      config,
		responsePkg: *response.NewResponse(options),
		AuthService: authService,
		JWTService:  jwtService,
	}
}

const (
	TestEmail        = "test1@e.com"
	TestPasswordHash = "$2a$10$rPoAoGvsaQZ/tRWf3DZphuUs1LbWIky0XppainCMrISRcDe8FOH0C"
	TestPassword     = "1"
	TestName         = "test"
)
