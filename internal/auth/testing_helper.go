package auth

import (
	"link-generator/internal/models"
	"link-generator/pkg/response"
)

type mockAccountService struct{}

func (m *mockAccountService) CreateAccount(userId uint, name, email string) (*models.Account, error) {
	return &models.Account{}, nil
}

func (m *mockAccountService) GetAccountInfoByEmail(email string) (*models.AccountInfo, error) {
	return &models.AccountInfo{Is2FAEnabled: false}, nil
}

func (m *mockAccountService) Setup2FA(email string) (string, error) { return "", nil }
func (m *mockAccountService) Verify2Fa(code, email string) bool     { return false }

// NewAuthHandlerForTest создает AuthHandler для тестов
func NewAuthHandlerForTest(authService *AuthService) *AuthHandler {
	headersMap := map[string]string{
		"Content-Type": "application/json",
	}

	options := &response.ResponseOptions{
		HeadersMap: headersMap,
	}

	return &AuthHandler{
		responsePkg:    *response.NewResponse(options),
		AuthService:    authService,
		AccountService: &mockAccountService{},
	}
}

const (
	TestEmail        = "test1@e.com"
	TestPasswordHash = "$2a$10$rPoAoGvsaQZ/tRWf3DZphuUs1LbWIky0XppainCMrISRcDe8FOH0C"
	TestPassword     = "1"
	TestName         = "test"
)
