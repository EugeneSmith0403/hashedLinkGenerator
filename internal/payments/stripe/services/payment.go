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

type CreatePaymentIntentParams struct {
	AccountId  uint
	UserId     uint
	CustomerID string
	CardType   string
	Currency   stripe.Currency
	Amount     int64
	PlanId     uint
}

func (s *PaymentService) CreatePaymentIntent(p CreatePaymentIntentParams) (*stripeGo.PaymentIntent, error) {
	paymentID := uuid.New()

	piParams := &stripeGo.PaymentIntentCreateParams{
		Customer:      stripe.String(p.CustomerID),
		PaymentMethod: stripe.String(p.CardType),
		Amount:        stripe.Int64(p.Amount * 100),
		Currency:      stripe.String(p.Currency),
		Metadata: map[string]string{
			"payment_id": paymentID.String(),
			"plan_id":    fmt.Sprintf("%d", p.PlanId),
			"user_id":    fmt.Sprintf("%d", p.UserId),
		},
	}

	pi, err := s.stripeProvider.V1PaymentIntents.Create(s.ctx, piParams)
	if err != nil {
		return nil, err
	}

	payment := paymentFromIntent(paymentID, p.AccountId, pi)
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
		if payment.PlanID == nil {
			if planIDStr, ok := pi.Metadata["plan_id"]; ok {
				if v, err := strconv.ParseUint(planIDStr, 10, 64); err == nil {
					uid := uint(v)
					payment.PlanID = &uid
				}
			}
		}
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

func (s *PaymentService) RefundPaymentIntent(paymentIntentID string) error {
	_, err := s.stripeProvider.V1Refunds.Create(s.ctx, &stripeGo.RefundCreateParams{
		PaymentIntent: stripe.String(paymentIntentID),
	})
	return err
}

func (s *PaymentService) CancelPaymentInDB(paymentIntentID string) error {
	p, err := s.paymentRepository.GetByPaymentIntentID(paymentIntentID)
	if err != nil || p == nil {
		return err
	}
	p.Status = stripeGo.PaymentIntentStatusCanceled
	_, err = s.paymentRepository.Save(p)
	return err
}

func (s *PaymentService) LinkSubscription(piID string, subscriptionID uint) error {
	return s.paymentRepository.LinkSubscriptionByPI(piID, subscriptionID)
}

func (s *PaymentService) DetectPaymentWebhook(payload []byte, sigHeader string) (*stripeGo.Event, error) {
	event, err := webhook.ConstructEvent(payload, sigHeader, s.webhookSecret)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
