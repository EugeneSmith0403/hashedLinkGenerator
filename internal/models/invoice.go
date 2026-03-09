package models

import stripeGo "github.com/stripe/stripe-go/v84"

type InvoiceEventType string

const (
	InvoiceCreated               InvoiceEventType = "invoice.created"
	InvoiceUpdated               InvoiceEventType = "invoice.updated"
	InvoiceDeleted               InvoiceEventType = "invoice.deleted"
	InvoiceFinalized             InvoiceEventType = "invoice.finalized"
	InvoicePaymentSucceeded      InvoiceEventType = "invoice.payment_succeeded"
	InvoicePaymentFailed         InvoiceEventType = "invoice.payment_failed"
	InvoicePaymentActionRequired InvoiceEventType = "invoice.payment_action_required"
	InvoiceVoided                InvoiceEventType = "invoice.voided"
	InvoiceMarkedUncollectible   InvoiceEventType = "invoice.marked_uncollectible"
	InvoiceSent                  InvoiceEventType = "invoice.sent"
	InvoiceUpcoming              InvoiceEventType = "invoice.upcoming"
)

type InvoiceMessage struct {
	EventType InvoiceEventType `json:"event_type"`
	Data      stripeGo.Invoice `json:"data"`
}
