package account

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"log"
	"time"

	"link-generator/internal/models"
	payments "link-generator/internal/payments/models"
	"link-generator/internal/user"
	pkgRedis "link-generator/pkg/redis"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAccountNotFound = errors.New("account not found")
)

const accountCacheTTL = 7 * 24 * time.Hour

type AccountServiceDeps struct {
	AccountRepository *AccountRepository
	PaymentService    payments.ICustomerAccountService
	UserRepository    *user.UserRepository
	Redis             *pkgRedis.Redis
}

type AccountService struct {
	AccountRepository *AccountRepository
	PaymentService    payments.ICustomerAccountService
	UserRepository    *user.UserRepository
	redis             *pkgRedis.Redis
}

func NewAccountService(accRep AccountServiceDeps) *AccountService {
	return &AccountService{
		AccountRepository: accRep.AccountRepository,
		PaymentService:    accRep.PaymentService,
		UserRepository:    accRep.UserRepository,
		redis:             accRep.Redis,
	}
}

func (s *AccountService) GetAccountByEmail(email string) (*Account, error) {
	if cached := s.redis.Get(accountCacheKey(email)); cached != "" {
		var account Account
		if err := json.Unmarshal([]byte(cached), &account); err == nil {
			return &account, nil
		}
	}

	foundUser, err := s.UserRepository.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if foundUser == nil {
		return nil, ErrUserNotFound
	}

	foundAccount, err := s.AccountRepository.FindByUserId(foundUser.Model.ID)
	if err != nil {
		return nil, err
	}
	if foundAccount == nil {
		return nil, ErrAccountNotFound
	}

	s.setAccountCache(email, foundAccount)

	return foundAccount, nil
}

func (rep *AccountService) UpdateAccount(userId uint, name, email string) (*Account, error) {
	account, err := rep.AccountRepository.FindByUserId(userId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	_, custErr := rep.PaymentService.UpdateCustomerAccount(account.CustomerID, name, email)
	if custErr != nil {
		return nil, custErr
	}

	rep.setAccountCache(email, account)

	return account, nil
}

func (rep *AccountService) CreateAccount(userId uint, name, email string) (*Account, error) {

	existedAccount, accErr := rep.AccountRepository.FindById(userId)

	if accErr != nil {
		return nil, accErr
	}

	if existedAccount != nil {
		return existedAccount, nil
	}

	customer, custErr := rep.PaymentService.CreateCustomerAccount(name, email)

	if custErr != nil {
		return nil, custErr
	}

	createdAccount, err := rep.AccountRepository.Create(&Account{
		CustomerID:    customer.ID,
		AccountStatus: StatusActive,
		Provider:      ProviderStripe,
		UserID:        userId,
	})

	if err != nil {
		return nil, err
	}

	rep.setAccountCache(email, createdAccount)

	return createdAccount, nil
}

func (s *AccountService) setAccountCache(email string, account *Account) {
	if data, err := json.Marshal(account); err == nil {
		s.redis.Set(accountCacheKey(email), string(data), accountCacheTTL)
	}
}

func accountCacheKey(email string) string {
	return fmt.Sprintf("account:%s", email)
}

func (s *AccountService) GetAccountInfoByEmail(email string) (*models.AccountInfo, error) {
	acc, err := s.GetAccountByEmail(email)
	if err != nil {
		return nil, err
	}
	return &models.AccountInfo{
		UserID:       acc.UserID,
		Is2FAEnabled: acc.Is2FAEnabled,
		TotpSecret:   acc.TotpSecret,
	}, nil
}

func (s *AccountService) Setup2FA(email string) (string, error) {
	foundUser, err := s.UserRepository.FindByEmail(email)
	if err != nil {
		return "", err
	}
	if foundUser == nil {
		return "", ErrUserNotFound
	}

	foundAccount, err := s.AccountRepository.FindByUserId(foundUser.Model.ID)
	if err != nil {
		return "", err
	}
	if foundAccount == nil {
		return "", ErrAccountNotFound
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "LinkShort",
		AccountName: email,
	})
	if err != nil {
		return "", err
	}

	foundAccount.TotpSecret = key.Secret()
	foundAccount.Is2FAEnabled = true
	if _, err = s.AccountRepository.Update(foundAccount); err != nil {
		return "", err
	}

	s.redis.Set(accountCacheKey(email), "", 1)

	img, err := key.Image(200, 200)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (s *AccountService) Verify2Fa(code, email string) bool {
	accInfo, accErr := s.GetAccountInfoByEmail(email)

	if accErr != nil {
		log.Printf("[2fa] GetAccountInfoByEmail error for %s: %v", email, accErr)
		return false
	}

	if accInfo.TotpSecret == "" {
		log.Printf("[2fa] TotpSecret is empty for %s — Setup2FA was not completed", email)
		return false
	}

	now := time.Now().UTC()
	expected, _ := totp.GenerateCode(accInfo.TotpSecret, now)
	log.Printf("[2fa] server time=%s expected=%s got=%s", now.Format(time.RFC3339), expected, code)

	isValid, _ := totp.ValidateCustom(code, accInfo.TotpSecret, now, totp.ValidateOpts{
		Period:    30,
		Skew:      5,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	log.Printf("[2fa] Verify for %s: secret_len=%d valid=%v", email, len(accInfo.TotpSecret), isValid)

	return isValid
}
