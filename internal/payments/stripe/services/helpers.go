package services

import (
	"encoding/json"
	"strconv"

	paymentmodels "link-generator/internal/payments/models"

	"github.com/google/uuid"
	stripeGo "github.com/stripe/stripe-go/v84"
	"gorm.io/datatypes"
)

func paymentFromIntent(id uuid.UUID, accountID uint, pi *stripeGo.PaymentIntent) *paymentmodels.Payment {
	metaJSON, _ := json.Marshal(pi.Metadata)
	failureCode, failureMessage := extractFailure(pi)
	var pmType string
	if len(pi.PaymentMethodTypes) > 0 {
		pmType = pi.PaymentMethodTypes[0]
	}
	p := &paymentmodels.Payment{
		ID:                id,
		AccountID:         accountID,
		PaymentIntentID:   pi.ID,
		ChargeID:          extractChargeID(pi),
		Amount:            pi.Amount,
		NetAmount:         pi.Amount,
		Currency:          string(pi.Currency),
		Status:            pi.Status,
		PaymentMethodType: pmType,
		FailureCode:       failureCode,
		FailureMessage:    failureMessage,
		ProviderMetadata:  datatypes.JSON(metaJSON),
	}
	if planIDStr, ok := pi.Metadata["plan_id"]; ok {
		if v, err := strconv.ParseUint(planIDStr, 10, 64); err == nil {
			uid := uint(v)
			p.PlanID = &uid
		}
	}
	return p
}

func extractChargeID(pi *stripeGo.PaymentIntent) *string {
	if pi.LatestCharge != nil {
		return &pi.LatestCharge.ID
	}
	return nil
}

func extractFailure(pi *stripeGo.PaymentIntent) (code, message string) {
	if pi.LastPaymentError != nil {
		return string(pi.LastPaymentError.Code), pi.LastPaymentError.Msg
	}
	return "", ""
}
