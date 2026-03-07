package invoice

import (
	"context"
	"fmt"

	paymentrepo "adv/go-http/internal/payments/payment"

	"github.com/google/uuid"
	stripeGo "github.com/stripe/stripe-go/v84"
)

type InvoiceServiceDeps struct {
	StripeClient      *stripeGo.Client
	InvoiceRepository *InvoiceRepository
	PaymentRepository *paymentrepo.PaymentRepository
}

type InvoiceService struct {
	stripeProvider    *stripeGo.Client
	invoiceRepository *InvoiceRepository
	paymentRepository *paymentrepo.PaymentRepository
	ctx               context.Context
}

func NewInvoiceService(deps InvoiceServiceDeps) *InvoiceService {
	return &InvoiceService{
		stripeProvider:    deps.StripeClient,
		invoiceRepository: deps.InvoiceRepository,
		paymentRepository: deps.PaymentRepository,
		ctx:               context.Background(),
	}
}

func (s *InvoiceService) CreatePaymentAndInvoiceFromStripeInvoice(inv *stripeGo.Invoice, accountID uint, subscriptionID *uint) error {
	saved, err := s.upsertInvoice(inv, accountID, subscriptionID)
	if err != nil {
		return err
	}

	retrieveParams := &stripeGo.InvoiceRetrieveParams{}
	retrieveParams.AddExpand("payments")
	fullInv, fetchErr := s.stripeProvider.V1Invoices.Retrieve(s.ctx, inv.ID, retrieveParams)
	if fetchErr != nil {
		return fmt.Errorf("fetch invoice payments: %w", fetchErr)
	}

	piID := piIDFromInvoice(fullInv)
	if piID == "" {
		return nil
	}

	if accountID == 0 {
		if p, lookupErr := s.paymentRepository.GetByPaymentIntentID(piID); lookupErr == nil && p != nil {
			accountID = p.AccountID
			if saved.AccountID == 0 && accountID != 0 {
				saved.AccountID = accountID
				if saved, err = s.invoiceRepository.Update(saved); err != nil {
					return fmt.Errorf("update invoice accountID: %w", err)
				}
			}
		}
	}

	return s.syncPaymentForInvoice(piID, accountID, saved, inv)
}

func (s *InvoiceService) CreateInitialInvoice(billingID string, accountID, subscriptionID uint) error {
	existing, err := s.invoiceRepository.GetByBillingID(billingID)
	if err != nil {
		return fmt.Errorf("check existing invoice: %w", err)
	}
	if existing != nil {
		return nil
	}
	subID := subscriptionID
	inv := &Invoice{
		AccountID:      accountID,
		SubscriptionID: &subID,
		BillingID:      billingID,
		Status:         InvoiceStatusOpen,
	}
	_, err = s.invoiceRepository.Create(inv)
	return err
}

func (s *InvoiceService) CreateInvoiceForPayment(pi *stripeGo.PaymentIntent) error {
	if pi == nil {
		return fmt.Errorf("CreateInvoiceForPayment: payment intent is nil")
	}

	parsedUUID, err := uuid.Parse(pi.Metadata["payment_id"])
	if err != nil {
		return fmt.Errorf("invalid payment_id in metadata: %w", err)
	}
	payment, err := s.paymentRepository.GetByUuid(parsedUUID)
	if err != nil {
		return fmt.Errorf("get payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found for intent: %s", pi.ID)
	}
	if payment.InvoiceID != nil {
		return nil
	}

	paid, err := s.createStripeInvoice(pi.Customer.ID, pi.Amount, string(pi.Currency))
	if err != nil {
		return err
	}

	saved, err := s.saveLocalInvoice(paid, payment.AccountID)
	if err != nil {
		return err
	}

	return s.paymentRepository.LinkInvoice(payment.ID, saved.ID)
}

