package subscription

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"link-generator/internal/payments/plan"
	"link-generator/internal/payments/payment"

	"github.com/stripe/stripe-go/v84"
	stripeGo "github.com/stripe/stripe-go/v84"
)

type SubscriptionServiceDeps struct {
	SubscriptionRepository *SubscriptionRepository
	PlanRepository         *plan.PlanRepository
	PaymentRepository      *payment.PaymentRepository
	StripeClient           *stripeGo.Client
	Ctx                    context.Context
}

type SubscriptionService struct {
	subscriptionRepository *SubscriptionRepository
	planRepository         *plan.PlanRepository
	paymentRepository      *payment.PaymentRepository
	stripeProvider         *stripeGo.Client
	ctx                    context.Context
}

func NewSubscriptionService(deps SubscriptionServiceDeps) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepository: deps.SubscriptionRepository,
		planRepository:         deps.PlanRepository,
		paymentRepository:      deps.PaymentRepository,
		stripeProvider:         deps.StripeClient,
		ctx:                    deps.Ctx,
	}
}

func (s *SubscriptionService) AddPaymentMethod(customerID string) (*stripeGo.SetupIntent, error) {
	return s.stripeProvider.V1SetupIntents.Create(s.ctx, &stripeGo.SetupIntentCreateParams{
		Customer: stripe.String(customerID),
		Usage:    stripe.String("off_session"),
	})
}

func (s *SubscriptionService) CreateSubscription(userID, planID uint, customedId, price, paymentMethodId string) (*stripeGo.Subscription, error) {
	p, err := s.planRepository.GetByID(planID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("plan %d not found", planID)
	}

	sub, err := s.stripeProvider.V1Subscriptions.Create(s.ctx, &stripeGo.SubscriptionCreateParams{
		Customer:             stripe.String(customedId),
		DefaultPaymentMethod: stripe.String(paymentMethodId),
		Items: []*stripeGo.SubscriptionCreateItemParams{
			{
				Price: stripe.String(price),
			},
		},
		Metadata: map[string]string{
			"user_id": fmt.Sprintf("%d", userID),
			"plan_id": fmt.Sprintf("%d", planID),
		},
	})
	if err != nil {
		return nil, err
	}

	start := time.Unix(sub.StartDate, 0)

	entity := &Subscription{
		UserID:             userID,
		PlanID:             planID,
		BillingID:          sub.ID,
		CustomerID:         customedId,
		Status:             SubscriptionStatus(sub.Status),
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   planPeriodEnd(p.IntervalMonths, start),
	}

	entity.CancelAt, entity.CanceledAt, entity.TrialStart, entity.TrialEnd = mapSubTimestamps(sub)

	if _, err = s.subscriptionRepository.Create(entity); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) CancelSubscription(subId string) (*Subscription, error) {
	canceledSub, err := s.stripeProvider.V1Subscriptions.Cancel(s.ctx, subId, nil)
	if err != nil {
		return nil, err
	}

	canceledAt := time.Now()
	sub, err := s.subscriptionRepository.UpdateByBillingId(canceledSub.ID, &SubscriptionUpdate{
		Status:     SubscriptionStatusCanceled,
		CanceledAt: &canceledAt,
	})
	if err != nil {
		return nil, err
	}

	if s.paymentRepository != nil {
		_ = s.paymentRepository.CancelBySubscriptionID(sub.ID)
	}

	return sub, nil
}

func (s *SubscriptionService) MarkCanceled(billingID string) (*Subscription, error) {
	canceledAt := time.Now()
	sub, err := s.subscriptionRepository.UpdateByBillingId(billingID, &SubscriptionUpdate{
		Status:     SubscriptionStatusCanceled,
		CanceledAt: &canceledAt,
	})
	if err != nil {
		return nil, err
	}

	if s.paymentRepository != nil {
		_ = s.paymentRepository.CancelBySubscriptionID(sub.ID)
	}

	return sub, nil
}

func (s *SubscriptionService) GetByBillingID(billingID string) (*Subscription, error) {
	return s.subscriptionRepository.GetByBillingID(billingID)
}

func (s *SubscriptionService) GetSubscriptionByUserId(userID uint) (*Subscription, error) {
	return s.subscriptionRepository.GetSubscriptionByUserId(userID)
}

func (s *SubscriptionService) HasActiveSubscription(userID uint) (bool, error) {
	sub, err := s.subscriptionRepository.GetActiveByUserID(userID)
	if err != nil {
		return false, err
	}
	return sub != nil, nil
}

func (s *SubscriptionService) UpdateSubscriptionFromEvent(sub *stripeGo.Subscription) error {
	existing, err := s.subscriptionRepository.GetByBillingID(sub.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	p, err := s.planRepository.GetByID(existing.PlanID)
	if err != nil {
		return fmt.Errorf("get plan: %w", err)
	}
	if p == nil {
		return fmt.Errorf("plan %d not found", existing.PlanID)
	}

	start := time.Unix(sub.StartDate, 0)
	upd := &SubscriptionUpdate{
		Status:             SubscriptionStatus(sub.Status),
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   planPeriodEnd(p.IntervalMonths, start),
	}

	upd.CancelAt, upd.CanceledAt, upd.TrialStart, upd.TrialEnd = mapSubTimestamps(sub)

	_, err = s.subscriptionRepository.Update(existing.ID, upd)
	return err
}

func (s *SubscriptionService) CreateFromStripeSub(sub *stripeGo.Subscription, userID, planID uint) (*Subscription, error) {
	p, err := s.planRepository.GetByID(planID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("plan %d not found", planID)
	}

	var customerID string
	if sub.Customer != nil {
		customerID = sub.Customer.ID
	}

	start := time.Unix(sub.StartDate, 0)
	entity := &Subscription{
		UserID:             userID,
		PlanID:             planID,
		BillingID:          sub.ID,
		CustomerID:         customerID,
		Status:             SubscriptionStatus(sub.Status),
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   planPeriodEnd(p.IntervalMonths, start),
	}
	entity.CancelAt, entity.CanceledAt, entity.TrialStart, entity.TrialEnd = mapSubTimestamps(sub)

	return s.subscriptionRepository.Create(entity)
}

func (s *SubscriptionService) CreateFromPaymentIntent(pi *stripeGo.PaymentIntent) (*Subscription, error) {
	userIDStr, ok := pi.Metadata["user_id"]
	if !ok {
		return nil, errors.New("missing user_id in payment intent metadata")
	}
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id in metadata: %w", err)
	}
	userID := uint(userID64)

	planIDStr, ok := pi.Metadata["plan_id"]
	if !ok {
		return nil, errors.New("missing plan_id in payment intent metadata")
	}
	planID64, err := strconv.ParseUint(planIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid plan_id in metadata: %w", err)
	}
	planID := uint(planID64)

	existing, err := s.subscriptionRepository.GetActiveByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	p, err := s.planRepository.GetByID(planID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("plan %d not found", planID)
	}

	var customerID string
	if pi.Customer != nil {
		customerID = pi.Customer.ID
	}

	start := time.Unix(pi.Created, 0)
	entity := &Subscription{
		UserID:             userID,
		PlanID:             planID,
		BillingID:          pi.ID,
		CustomerID:         customerID,
		Status:             SubscriptionStatusActive,
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   planPeriodEnd(p.IntervalMonths, start),
	}

	return s.subscriptionRepository.Create(entity)
}
