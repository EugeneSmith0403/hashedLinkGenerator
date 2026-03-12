package consts

const (
	PaymentIntentExchange = "paymentIntent"
	PaymentIntentRouting  = "PaymentIntent.succeed"
	PaymentIntentQueue    = "paymentIntentQueue"
	SubscriptionExchange  = "Subscription"
	SubscriptionRouting   = "Subscription.succeed"
	SubscriptionQueue     = "subscriptionQueue"
	InvoiceExchange       = "Invoice"
	InvoiceRouting        = "Invoice.payment_succeeded"
	InvoiceQueue          = "invoiceQueue"
	StatsExchange         = "Stats"
	StatsRouting          = "Stats.link_visited"
	StatsQueue            = "statsQueue"
)
