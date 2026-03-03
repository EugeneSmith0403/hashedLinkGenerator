package payments

import "github.com/stripe/stripe-go/v84"

type IPaymentService interface {
	CreateCustomerAccount(name, email string) (*stripe.Customer, error)
	UpdateCustomerAccount(customerID, name, email string) (*stripe.Customer, error)
	CreatePaymentIntent(accountId uint, customerID string, cardType string, currency stripe.Currency, amount int64, planId uint) (*stripe.PaymentIntent, error)
	CancelPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error)
	DetectPaymentWebhook(payload []byte, sigHeader string) (*stripe.Event, error)
}
