package account

import (
	"errors"
	"fmt"

	payments "adv/go-http/internal/payments/models"
	"adv/go-http/internal/user"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAccountNotFound = errors.New("account not found")
)

type AccountServiceDeps struct {
	AccountRepository *AccountRepository
	PaymentService    payments.ICustomerAccountService
	UserRepository    *user.UserRepository
}

type AccountService struct {
	AccountRepository *AccountRepository
	PaymentService    payments.ICustomerAccountService
	UserRepository    *user.UserRepository
}

func NewAccountService(accRep AccountServiceDeps) *AccountService {
	return &AccountService{
		AccountRepository: accRep.AccountRepository,
		PaymentService:    accRep.PaymentService,
		UserRepository:    accRep.UserRepository,
	}
}

func (s *AccountService) GetAccountByEmail(email string) (*Account, error) {
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

	return foundAccount, nil
}

func (rep *AccountService) UpdateAccount(userId uint, name, email string) (*Account, error) {
	account, err := rep.AccountRepository.FindByUserId(userId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("account not found")
	}

	_, custErr := rep.PaymentService.UpdateCustomerAccount(account.CustomerID, name, email)
	if custErr != nil {
		return nil, custErr
	}

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

	return createdAccount, nil
}
