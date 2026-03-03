package stripe

import "github.com/google/uuid"

type ConfirmPaymentIntentResponse struct {
	Confirmed    bool      `json:"confirmed"`
	PaymentId    uuid.UUID `json:"paymentId"`
	ConfirmedUrl string    `json:"confirmedUrl"`
}
