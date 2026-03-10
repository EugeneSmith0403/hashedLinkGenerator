package main

import (
	"context"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"

	"link-generator/cmd/shared"
	"link-generator/internal/consts"
	piConsumer "link-generator/internal/consumers"
	"link-generator/internal/payments/invoice"
	"link-generator/internal/payments/payment"
	"link-generator/internal/payments/plan"
	stripeServices "link-generator/internal/payments/stripe/services"
	"link-generator/internal/payments/subscription"
	"link-generator/pkg/db"
	rabbitmq "link-generator/pkg/rabbitMq"
)

func main() {
	cfg := shared.LoadConfigs()

	database := db.NewDb(cfg)
	defer database.Close()

	rabbitMq := rabbitmq.NewRabbitMq(cfg.RabbitMq)
	defer rabbitMq.Close()

	stripeClient := stripeGo.NewClient(cfg.Stripe.ApiKey)

	paymentRepo := payment.NewPaymentRepository(database)
	invoiceRepo := invoice.NewInvoiceRepository(database)
	subscriptionRepo := subscription.NewSubscriptionRepository(database)
	planRepo := plan.NewPlanRepository(database)

	consumer := piConsumer.NewPaymentIntentConsumer(piConsumer.PaymentIntentConsumerDeps{
		PaymentSvc: stripeServices.NewPaymentService(stripeServices.PaymentServiceDeps{
			StripeClient:      stripeClient,
			WebhookSecret:     cfg.Stripe.WebhookSecret,
			ReturnURL:         cfg.Stripe.ReturnURL,
			PaymentRepository: paymentRepo,
		}),
		InvoiceSvc: invoice.NewInvoiceService(invoice.InvoiceServiceDeps{
			StripeClient:           stripeClient,
			InvoiceRepository:      invoiceRepo,
			PaymentRepository:      paymentRepo,
			SubscriptionRepository: subscriptionRepo,
		}),
		SubscriptionSvc: subscription.NewSubscriptionService(subscription.SubscriptionServiceDeps{
			SubscriptionRepository: subscriptionRepo,
			PlanRepository:         planRepo,
			PaymentRepository:      paymentRepo,
			StripeClient:           stripeClient,
			Ctx:                    context.Background(),
		}),
	})

	msgs, err := rabbitMq.CreateConsumer(&rabbitmq.ConsumerOptions{
		Queue:   consts.PaymentIntentQueue,
		AutoAck: false,
	})
	if err != nil {
		log.Fatalf("[consumer] failed to create consumer: %v", err)
	}

	shared.RunConsumerLoop(msgs, consumer.Handle, "payment_intent")
}
