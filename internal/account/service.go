package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	payments "adv/go-http/internal/payments/models"
	"adv/go-http/internal/user"
	pkgRedis "adv/go-http/pkg/redis"
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
