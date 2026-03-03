package invoice

import (
	"encoding/json"
	"fmt"
	"time"

	paymentmodels "adv/go-http/internal/payments/models"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	stripeGo "github.com/stripe/stripe-go/v84"
	"gorm.io/datatypes"
)

func (s *InvoiceService) upsertInvoice(inv *stripeGo.Invoice, accountID uint, subscriptionID *uint) (*Invoice, error) {
	existing, err := s.invoiceRepository.GetByBillingID(inv.ID)
	if err != nil {
		return nil, fmt.Errorf("check existing invoice: %w", err)
	}

	metaJSON, _ := json.Marshal(inv)
	paidAt := paidAtFromTransitions(inv.StatusTransitions)

	if existing != nil {
		existing.Status = InvoiceStatus(inv.Status)
		existing.AmountDue = inv.AmountDue
		existing.AmountPaid = inv.AmountPaid
		existing.AmountRemaining = inv.AmountRemaining
		existing.HostedInvoiceURL = inv.HostedInvoiceURL
		existing.InvoicePDF = inv.InvoicePDF
		existing.PaidAt = paidAt
		existing.ProviderMetadata = datatypes.JSON(metaJSON)
		saved, err := s.invoiceRepository.Update(existing)
		if err != nil {
			return nil, fmt.Errorf("update invoice: %w", err)
		}
		return saved, nil
	}

	localInv := &Invoice{
		AccountID:        accountID,
		SubscriptionID:   subscriptionID,
		BillingID:        inv.ID,
		Status:           InvoiceStatus(inv.Status),
		AmountDue:        inv.AmountDue,
		AmountPaid:       inv.AmountPaid,
		AmountRemaining:  inv.AmountRemaining,
		Currency:         string(inv.Currency),
		HostedInvoiceURL: inv.HostedInvoiceURL,
		InvoicePDF:       inv.InvoicePDF,
		PaidAt:           paidAt,
		ProviderMetadata: datatypes.JSON(metaJSON),
	}
	saved, err := s.invoiceRepository.Create(localInv)
	if err != nil {
		return nil, fmt.Errorf("save invoice: %w", err)
	}
	return saved, nil
}

func (s *InvoiceService) syncPaymentForInvoice(piID string, accountID uint, savedInv *Invoice, inv *stripeGo.Invoice) error {
	existingPayment, err := s.paymentRepository.GetByPaymentIntentID(piID)
	if err != nil {
		return fmt.Errorf("check existing payment: %w", err)
	}
	if existingPayment != nil {
		if existingPayment.InvoiceID == nil {
			return s.paymentRepository.LinkInvoice(existingPayment.ID, savedInv.ID)
		}
		return nil
	}

	metaPayJSON, _ := json.Marshal(inv.Metadata)
	payment := &paymentmodels.Payment{
		ID:               uuid.New(),
		AccountID:        accountID,
		InvoiceID:        &savedInv.ID,
		PaymentIntentID:  piID,
		Amount:           inv.AmountPaid,
		NetAmount:        inv.AmountPaid,
		Currency:         string(inv.Currency),
		Status:           stripe.PaymentIntentStatusSucceeded,
		ProviderMetadata: datatypes.JSON(metaPayJSON),
	}
	_, err = s.paymentRepository.Create(payment)
	return err
}

func (s *InvoiceService) createStripeInvoice(customerID string, amount int64, currency string) (*stripeGo.Invoice, error) {
	_, err := s.stripeProvider.V1InvoiceItems.Create(s.ctx, &stripeGo.InvoiceItemCreateParams{
		Customer: stripe.String(customerID),
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice item: %w", err)
	}

	inv, err := s.stripeProvider.V1Invoices.Create(s.ctx, &stripeGo.InvoiceCreateParams{
		Customer:    stripe.String(customerID),
		AutoAdvance: stripe.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	finalized, err := s.stripeProvider.V1Invoices.FinalizeInvoice(s.ctx, inv.ID, &stripeGo.InvoiceFinalizeInvoiceParams{
		AutoAdvance: stripe.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("finalize invoice: %w", err)
	}

	paid, err := s.stripeProvider.V1Invoices.Pay(s.ctx, finalized.ID, &stripeGo.InvoicePayParams{
		PaidOutOfBand: stripe.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("mark invoice paid: %w", err)
	}

	return paid, nil
}

func (s *InvoiceService) saveLocalInvoice(paid *stripeGo.Invoice, accountID uint) (*Invoice, error) {
	metaJSON, _ := json.Marshal(paid)
	localInv := &Invoice{
		AccountID:        accountID,
		BillingID:        paid.ID,
		Status:           InvoiceStatusPaid,
		AmountDue:        paid.AmountDue,
		AmountPaid:       paid.AmountPaid,
		AmountRemaining:  paid.AmountRemaining,
		Currency:         string(paid.Currency),
		HostedInvoiceURL: paid.HostedInvoiceURL,
		InvoicePDF:       paid.InvoicePDF,
		PaidAt:           paidAtFromTransitions(paid.StatusTransitions),
		ProviderMetadata: datatypes.JSON(metaJSON),
	}
	saved, err := s.invoiceRepository.Create(localInv)
	if err != nil {
		return nil, fmt.Errorf("save invoice: %w", err)
	}
	return saved, nil
}

func paidAtFromTransitions(t *stripeGo.InvoiceStatusTransitions) *time.Time {
	if t != nil && t.PaidAt != 0 {
		paidAt := time.Unix(t.PaidAt, 0)
		return &paidAt
	}
	return nil
}

func piIDFromInvoice(inv *stripeGo.Invoice) string {
	if inv.Payments == nil {
		return ""
	}
	for _, p := range inv.Payments.Data {
		if p.Payment != nil &&
			p.Payment.Type == stripeGo.InvoicePaymentPaymentTypePaymentIntent &&
			p.Payment.PaymentIntent != nil {
			return p.Payment.PaymentIntent.ID
		}
	}
	return ""
}
