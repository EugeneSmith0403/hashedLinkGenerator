package auth

import (
	"link-generator/pkg/response"
)

// NewAuthHandlerForTest создает AuthHandler для тестов
func NewAuthHandlerForTest(authService *AuthService) *AuthHandler {
	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}

	return &AuthHandler{
		responsePkg: *response.NewResponse(options),
		AuthService: authService,
	}
}

const (
	TestEmail        = "test1@e.com"
	TestPasswordHash = "$2a$10$rPoAoGvsaQZ/tRWf3DZphuUs1LbWIky0XppainCMrISRcDe8FOH0C"
	TestPassword     = "1"
	TestName         = "test"
)
