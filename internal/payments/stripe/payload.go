package stripe

import (
	"github.com/google/uuid"
)

type PaymentIntentPayload struct {
	CardType string `json:"cardType" validate:"required"`
	PlanId   uint   `json:"planId"   validate:"required,oneof=1 2 3"`
}

type ConfirmPaymentIntentPayload struct {
	PaymentId uuid.UUID `json:"paymentId" validate:"required"`
}
