package services

import (
	"context"

	"github.com/stripe/stripe-go/v84"
	stripeGo "github.com/stripe/stripe-go/v84"
)

type CustomerAccountServiceDeps struct {
	StripeClient *stripeGo.Client
}

type CustomerAccountService struct {
	stripeProvider *stripeGo.Client
	ctx            context.Context
}

func NewCustomerAccountService(deps CustomerAccountServiceDeps) *CustomerAccountService {
	return &CustomerAccountService{
		stripeProvider: deps.StripeClient,
		ctx:            context.Background(),
	}
}

func (s *CustomerAccountService) CreateCustomerAccount(name, email string) (*stripeGo.Customer, error) {
	params := &stripeGo.CustomerCreateParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}
	result, err := s.stripeProvider.V1Customers.Create(s.ctx, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *CustomerAccountService) UpdateCustomerAccount(customerID, name, email string) (*stripeGo.Customer, error) {
	params := &stripeGo.CustomerUpdateParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}
	result, err := s.stripeProvider.V1Customers.Update(s.ctx, customerID, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *CustomerAccountService) SetDefaultPaymentMethod(customerID, paymentMethodID string) error {
	_, err := s.stripeProvider.V1Customers.Update(s.ctx, customerID, &stripeGo.CustomerUpdateParams{
		InvoiceSettings: &stripeGo.CustomerUpdateInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	})
	return err
}
