package main

import (
	"context"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"

	"link-generator/cmd/shared"
	"link-generator/internal/account"
	"link-generator/internal/consts"
	subConsumer "link-generator/internal/consumers"
	"link-generator/internal/payments/invoice"
	"link-generator/internal/payments/payment"
	"link-generator/internal/payments/plan"
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

	subscriptionRepo := subscription.NewSubscriptionRepository(database)
	planRepo := plan.NewPlanRepository(database)
	accountRepo := account.NewAccountRepository(database)
	invoiceRepo := invoice.NewInvoiceRepository(database)
	paymentRepo := payment.NewPaymentRepository(database)

	consumer := subConsumer.NewSubscriptionConsumer(&subConsumer.SubscriptionConsumerDeps{
		SubscriptionSvc: subscription.NewSubscriptionService(subscription.SubscriptionServiceDeps{
			SubscriptionRepository: subscriptionRepo,
			PlanRepository:         planRepo,
			PaymentRepository:      paymentRepo,
			StripeClient:           stripeClient,
			Ctx:                    context.Background(),
		}),
		InvoiceSvc: invoice.NewInvoiceService(invoice.InvoiceServiceDeps{
			StripeClient:           stripeClient,
			InvoiceRepository:      invoiceRepo,
			PaymentRepository:      paymentRepo,
			SubscriptionRepository: subscriptionRepo,
		}),
		AccountRepository: accountRepo,
	})

	msgs, err := rabbitMq.CreateConsumer(&rabbitmq.ConsumerOptions{
		Queue:   consts.SubscriptionQueue,
		AutoAck: false,
	})
	if err != nil {
		log.Fatalf("[consumer] failed to create consumer: %v", err)
	}

	shared.RunConsumerLoop(msgs, consumer.Handle, "subscription")
}
