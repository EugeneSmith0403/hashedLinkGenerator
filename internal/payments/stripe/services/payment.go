package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	paymentrepo "adv/go-http/internal/payments/payment"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	stripeGo "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
	"gorm.io/datatypes"
)

type PaymentServiceDeps struct {
	StripeClient      *stripeGo.Client
	WebhookSecret     string
	ReturnURL         string
	PaymentRepository *paymentrepo.PaymentRepository
}

type PaymentService struct {
	stripeProvider    *stripeGo.Client
	webhookSecret     string
	returnURL         string
	ctx               context.Context
	paymentRepository *paymentrepo.PaymentRepository
}

func NewPaymentService(deps PaymentServiceDeps) *PaymentService {
	return &PaymentService{
		stripeProvider:    deps.StripeClient,
		webhookSecret:     deps.WebhookSecret,
		returnURL:         deps.ReturnURL,
		ctx:               context.Background(),
		paymentRepository: deps.PaymentRepository,
	}
}

type ConfirmPaymentIntentResponse struct {
	Confirmed    bool      `json:"confirmed"`
	PaymentId    uuid.UUID `json:"paymentId"`
	ConfirmedUrl string    `json:"confirmedUrl"`
}

func (s *PaymentService) CreatePaymentIntent(
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

	payment := paymentFromIntent(paymentID, accountId, pi)
	if _, saveErr := s.paymentRepository.Create(payment); saveErr != nil {
		return nil, saveErr
	}

	return pi, nil
}

func (s *PaymentService) ConfirmPaymentIntent(paymentId uuid.UUID) (*ConfirmPaymentIntentResponse, error) {
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

func (s *PaymentService) UpdatePaymentFromIntent(pi *stripeGo.PaymentIntent) (*stripeGo.PaymentIntent, error) {
	parsedUUID, parseErr := uuid.Parse(pi.Metadata["payment_id"])
	if parseErr != nil {
		return nil, fmt.Errorf("invalid payment_id in metadata: %w", parseErr)
	}

	payment, err := s.paymentRepository.GetByUuid(parsedUUID)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		var accountID uint
		if uid, e := strconv.ParseUint(pi.Metadata["user_id"], 10, 64); e == nil {
			accountID = uint(uid)
		}
		payment = paymentFromIntent(parsedUUID, accountID, pi)
	} else {
		metaJSON, _ := json.Marshal(pi.Metadata)
		payment.Status = pi.Status
		payment.ChargeID = extractChargeID(pi)
		payment.FailureCode, payment.FailureMessage = extractFailure(pi)
		payment.ProviderMetadata = datatypes.JSON(metaJSON)
	}

	_, err = s.paymentRepository.Save(payment)
	return pi, err
}

func (s *PaymentService) CancelPaymentIntent(paymentIntentID string) (*stripeGo.PaymentIntent, error) {
	result, err := s.stripeProvider.V1PaymentIntents.Cancel(s.ctx, paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PaymentService) DetectPaymentWebhook(payload []byte, sigHeader string) (*stripeGo.Event, error) {
	event, err := webhook.ConstructEvent(payload, sigHeader, s.webhookSecret)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

