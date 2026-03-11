package models

type AccountInfo struct {
	UserID       uint
	Is2FAEnabled bool
	TotpSecret   string
}

type IAccountService interface {
	GetAccountInfoByEmail(email string) (*AccountInfo, error)
	Setup2FA(email string) (string, error)
	Verify2Fa(code, email string) bool
}

type IAuthService interface {
	GenerateToken(email string) (string, error)
}
