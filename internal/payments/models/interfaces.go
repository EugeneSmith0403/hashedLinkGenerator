package payments

import "github.com/stripe/stripe-go/v84"

type ICustomerAccountService interface {
	CreateCustomerAccount(name, email string) (*stripe.Customer, error)
	UpdateCustomerAccount(customerID, name, email string) (*stripe.Customer, error)
}
