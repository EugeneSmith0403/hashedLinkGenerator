package models

import stripeGo "github.com/stripe/stripe-go/v84"

type PaymentIntentEventType string

const (
	PaymentIntentCreated                 PaymentIntentEventType = "payment_intent.created"
	PaymentIntentProcessing              PaymentIntentEventType = "payment_intent.processing"
	PaymentIntentSucceeded               PaymentIntentEventType = "payment_intent.succeeded"
	PaymentIntentPaymentFailed           PaymentIntentEventType = "payment_intent.payment_failed"
	PaymentIntentCanceled                PaymentIntentEventType = "payment_intent.canceled"
	PaymentIntentAmountCapturableUpdated PaymentIntentEventType = "payment_intent.amount_capturable_updated"
	PaymentIntentRequiresAction          PaymentIntentEventType = "payment_intent.requires_action"
	PaymentIntentPartiallyFunded         PaymentIntentEventType = "payment_intent.partially_funded"
)

type PaymentIntentMessage struct {
	EventType PaymentIntentEventType `json:"event_type"`
	Data      stripeGo.PaymentIntent `json:"data"`
}
