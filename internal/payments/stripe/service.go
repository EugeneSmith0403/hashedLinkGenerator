package stripe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	invoicemodel "adv/go-http/internal/payments/invoice"
	paymentmodels "adv/go-http/internal/payments/models"
	paymentrepo "adv/go-http/internal/payments/payment"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	stripeGo "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
	"gorm.io/datatypes"
)

type StripeDeps struct {
	StripeClient      *stripeGo.Client
	WebhookSecret     string
	ReturnURL         string
	PaymentRepository *paymentrepo.PaymentRepository
	InvoiceRepository *invoicemodel.InvoiceRepository
}

type StripeService struct {
	stripeProvider    *stripeGo.Client
	webhookSecret     string
	returnURL         string
	ctx               context.Context
	paymentRepository *paymentrepo.PaymentRepository
	invoiceRepository *invoicemodel.InvoiceRepository
}

func NewStripeService(deps StripeDeps) *StripeService {
	ctx := context.Background()
	return &StripeService{
		stripeProvider:    deps.StripeClient,
		webhookSecret:     deps.WebhookSecret,
		returnURL:         deps.ReturnURL,
		ctx:               ctx,
		paymentRepository: deps.PaymentRepository,
		invoiceRepository: deps.InvoiceRepository,
	}
}

func (s *StripeService) CreateCustomerAccount(name, email string) (*stripeGo.Customer, error) {
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

func (s *StripeService) UpdateCustomerAccount(customerID, name, email string) (*stripeGo.Customer, error) {
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

func (s *StripeService) SetDefaultPaymentMethod(customerID, paymentMethodID string) error {
	_, err := s.stripeProvider.V1Customers.Update(s.ctx, customerID, &stripeGo.CustomerUpdateParams{
		InvoiceSettings: &stripeGo.CustomerUpdateInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	})
	return err
}

func (s *StripeService) CreatePaymentIntent(
	accountId uint,
	customerID string,
	cardType string,
	currency stripe.Currency,
	amount int64,
	planId uint,
) (*stripeGo.PaymentIntent, error) {
	paymentID := uuid.New()

	params := &stripeGo.PaymentIntentCreateParams{
		Customer:      stripe.String(customerID),
		PaymentMethod: stripe.String(cardType),
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(currency),
		Metadata: map[string]string{
			"payment_id": paymentID.String(),
			"plan_id":    fmt.Sprintf("%d", planId),
			"user_id":    fmt.Sprintf("%d", accountId),
		},
	}

	pi, err := s.stripeProvider.V1PaymentIntents.Create(s.ctx, params)
	if err != nil {
		return nil, err
	}

	var chargeID *string
	if pi.LatestCharge != nil {
		chargeID = &pi.LatestCharge.ID
	}

	var pmType string
	if len(pi.PaymentMethodTypes) > 0 {
		pmType = pi.PaymentMethodTypes[0]
	}

	var failureCode, failureMessage string
	if pi.LastPaymentError != nil {
		failureCode = string(pi.LastPaymentError.Code)
		failureMessage = pi.LastPaymentError.Msg
	}

	metaJSON, _ := json.Marshal(pi.Metadata)

	payment := &paymentmodels.Payment{
		ID:                paymentID,
		AccountID:         accountId,
		PaymentIntentID:   pi.ID,
		ChargeID:          chargeID,
		Amount:            pi.Amount,
		NetAmount:         pi.Amount,
		Currency:          string(pi.Currency),
		Status:            pi.Status,
		PaymentMethodType: pmType,
		FailureCode:       failureCode,
		FailureMessage:    failureMessage,
		ProviderMetadata:  datatypes.JSON(metaJSON),
	}

	if _, saveErr := s.paymentRepository.Create(payment); saveErr != nil {
		return nil, saveErr
	}

	return pi, nil
}

func (s *StripeService) ConfirmPaymentIntent(paymentId uuid.UUID) (*ConfirmPaymentIntentResponse, error) {
	existedPayment, paymentErr := s.paymentRepository.GetByUuid(paymentId)

	if paymentErr != nil {
		return nil, paymentErr
	}

	if existedPayment == nil {
		return nil, errors.New("Payment intent is not found")
	}

	confirmed, err := s.stripeProvider.V1PaymentIntents.Confirm(s.ctx, existedPayment.PaymentIntentID, &stripeGo.PaymentIntentConfirmParams{
		ReturnURL: stripe.String(s.returnURL),
	})

	if err != nil {
		return nil, err
	}

	existedPayment.Status = confirmed.Status
	s.paymentRepository.Save(existedPayment)

	return &ConfirmPaymentIntentResponse{
		Confirmed:    *stripe.Bool(true),
		PaymentId:    paymentId,
		ConfirmedUrl: s.returnURL,
	}, nil

}

func (s *StripeService) UpdatePaymentFromIntent(pi *stripeGo.PaymentIntent) (*stripeGo.PaymentIntent, error) {
	parsedUUID, parseErr := uuid.Parse(pi.Metadata["payment_id"])
	if parseErr != nil {
		return nil, fmt.Errorf("invalid payment_id in metadata: %w", parseErr)
	}

	var chargeID *string
	if pi.LatestCharge != nil {
		chargeID = &pi.LatestCharge.ID
	}

	var failureCode, failureMessage string
	if pi.LastPaymentError != nil {
		failureCode = string(pi.LastPaymentError.Code)
		failureMessage = pi.LastPaymentError.Msg
	}

	metaJSON, _ := json.Marshal(pi.Metadata)

	payment, err := s.paymentRepository.GetByUuid(parsedUUID)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		var accountID uint
		if uid, e := strconv.ParseUint(pi.Metadata["user_id"], 10, 64); e == nil {
			accountID = uint(uid)
		}

		var pmType string
		if len(pi.PaymentMethodTypes) > 0 {
			pmType = pi.PaymentMethodTypes[0]
		}

		payment = &paymentmodels.Payment{
			ID:                parsedUUID,
			AccountID:         accountID,
			PaymentIntentID:   pi.ID,
			ChargeID:          chargeID,
			Amount:            pi.Amount,
			NetAmount:         pi.Amount,
			Currency:          string(pi.Currency),
			Status:            pi.Status,
			PaymentMethodType: pmType,
			FailureCode:       failureCode,
			FailureMessage:    failureMessage,
			ProviderMetadata:  datatypes.JSON(metaJSON),
		}
	} else {
		payment.Status = pi.Status
		payment.ChargeID = chargeID
		payment.FailureCode = failureCode
		payment.FailureMessage = failureMessage
		payment.ProviderMetadata = datatypes.JSON(metaJSON)
	}

	_, err = s.paymentRepository.Save(payment)
	return pi, err
}

func (s *StripeService) CancelPaymentIntent(paymentIntentID string) (*stripeGo.PaymentIntent, error) {
	result, err := s.stripeProvider.V1PaymentIntents.Cancel(s.ctx, paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *StripeService) DetectPaymentWebhook(payload []byte, sigHeader string) (*stripeGo.Event, error) {
	event, err := webhook.ConstructEvent(payload, sigHeader, s.webhookSecret)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *StripeService) CreatePaymentAndInvoiceFromStripeInvoice(inv *stripeGo.Invoice, accountID uint, subscriptionID *uint) error {
	existing, err := s.invoiceRepository.GetByBillingID(inv.ID)
	if err != nil {
		return fmt.Errorf("check existing invoice: %w", err)
	}

	var paidAt *time.Time
	if inv.StatusTransitions != nil && inv.StatusTransitions.PaidAt != 0 {
		t := time.Unix(inv.StatusTransitions.PaidAt, 0)
		paidAt = &t
	}

	metaJSON, _ := json.Marshal(inv)

	var saved *invoicemodel.Invoice
	if existing != nil {
		existing.Status = invoicemodel.InvoiceStatus(inv.Status)
		existing.AmountDue = inv.AmountDue
		existing.AmountPaid = inv.AmountPaid
		existing.AmountRemaining = inv.AmountRemaining
		existing.HostedInvoiceURL = inv.HostedInvoiceURL
		existing.InvoicePDF = inv.InvoicePDF
		existing.PaidAt = paidAt
		existing.ProviderMetadata = datatypes.JSON(metaJSON)
		saved, err = s.invoiceRepository.Update(existing)
		if err != nil {
			return fmt.Errorf("update invoice: %w", err)
		}
	} else {
		localInv := &invoicemodel.Invoice{
			AccountID:        accountID,
			SubscriptionID:   subscriptionID,
			BillingID:        inv.ID,
			Status:           invoicemodel.InvoiceStatus(inv.Status),
			AmountDue:        inv.AmountDue,
			AmountPaid:       inv.AmountPaid,
			AmountRemaining:  inv.AmountRemaining,
			Currency:         string(inv.Currency),
			HostedInvoiceURL: inv.HostedInvoiceURL,
			InvoicePDF:       inv.InvoicePDF,
			PaidAt:           paidAt,
			ProviderMetadata: datatypes.JSON(metaJSON),
		}
		saved, err = s.invoiceRepository.Create(localInv)
		if err != nil {
			return fmt.Errorf("save invoice: %w", err)
		}
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

	existingPayment, err := s.paymentRepository.GetByPaymentIntentID(piID)
	if err != nil {
		return fmt.Errorf("check existing payment: %w", err)
	}
	if existingPayment != nil {
		if existingPayment.InvoiceID == nil {
			return s.paymentRepository.LinkInvoice(existingPayment.ID, saved.ID)
		}
		return nil
	}

	metaPayJSON, _ := json.Marshal(inv.Metadata)
	payment := &paymentmodels.Payment{
		ID:               uuid.New(),
		AccountID:        accountID,
		InvoiceID:        &saved.ID,
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

func (s *StripeService) CreateInitialInvoice(billingID string, accountID, subscriptionID uint) error {
	existing, err := s.invoiceRepository.GetByBillingID(billingID)
	if err != nil {
		return fmt.Errorf("check existing invoice: %w", err)
	}
	if existing != nil {
		return nil
	}
	subID := subscriptionID
	inv := &invoicemodel.Invoice{
		AccountID:      accountID,
		SubscriptionID: &subID,
		BillingID:      billingID,
		Status:         invoicemodel.InvoiceStatusOpen,
	}
	_, err = s.invoiceRepository.Create(inv)
	return err
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

func (s *StripeService) CreateInvoiceForPayment(pi *stripeGo.PaymentIntent) error {
	if pi == nil {
		return errors.New("CreateInvoiceForPayment: payment intent is nil")
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
		return errors.New("payment not found for intent: " + pi.ID)
	}

	customerID := pi.Customer.ID

	_, err = s.stripeProvider.V1InvoiceItems.Create(s.ctx, &stripeGo.InvoiceItemCreateParams{
		Customer: stripe.String(customerID),
		Amount:   stripe.Int64(pi.Amount),
		Currency: stripe.String(string(pi.Currency)),
	})
	if err != nil {
		return fmt.Errorf("create invoice item: %w", err)
	}

	inv, err := s.stripeProvider.V1Invoices.Create(s.ctx, &stripeGo.InvoiceCreateParams{
		Customer:    stripe.String(customerID),
		AutoAdvance: stripe.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("create invoice: %w", err)
	}

	finalized, err := s.stripeProvider.V1Invoices.FinalizeInvoice(s.ctx, inv.ID, &stripeGo.InvoiceFinalizeInvoiceParams{
		AutoAdvance: stripe.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("finalize invoice: %w", err)
	}

	paid, err := s.stripeProvider.V1Invoices.Pay(s.ctx, finalized.ID, &stripeGo.InvoicePayParams{
		PaidOutOfBand: stripe.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("mark invoice paid: %w", err)
	}

	var paidAt *time.Time
	if paid.StatusTransitions != nil && paid.StatusTransitions.PaidAt != 0 {
		t := time.Unix(paid.StatusTransitions.PaidAt, 0)
		paidAt = &t
	}

	metaJSON, _ := json.Marshal(paid)
	localInv := &invoicemodel.Invoice{
		AccountID:        payment.AccountID,
		BillingID:        paid.ID,
		Status:           invoicemodel.InvoiceStatusPaid,
		AmountDue:        paid.AmountDue,
		AmountPaid:       paid.AmountPaid,
		AmountRemaining:  paid.AmountRemaining,
		Currency:         string(paid.Currency),
		HostedInvoiceURL: paid.HostedInvoiceURL,
		InvoicePDF:       paid.InvoicePDF,
		PaidAt:           paidAt,
		ProviderMetadata: datatypes.JSON(metaJSON),
	}

	saved, err := s.invoiceRepository.Create(localInv)
	if err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	return s.paymentRepository.LinkInvoice(payment.ID, saved.ID)
}
