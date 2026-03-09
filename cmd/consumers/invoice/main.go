package main

import (
	"context"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"

	"adv/go-http/cmd/shared"
	"adv/go-http/internal/account"
	"adv/go-http/internal/consts"
	invConsumer "adv/go-http/internal/consumers"
	"adv/go-http/internal/payments/invoice"
	"adv/go-http/internal/payments/payment"
	"adv/go-http/internal/payments/plan"
	"adv/go-http/internal/payments/subscription"
	"adv/go-http/internal/publishers"
	"adv/go-http/pkg/db"
	rabbitmq "adv/go-http/pkg/rabbitMq"
)

func main() {
	cfg := shared.LoadConfigs()

	database := db.NewDb(cfg)
	defer database.Close()

	rabbitMq := rabbitmq.NewRabbitMq(cfg.RabbitMq)
	defer rabbitMq.Close()

	publishers.NewInvoicePublisher(rabbitMq).CreateExchangeAndQueue()

	stripeClient := stripeGo.NewClient(cfg.Stripe.ApiKey)

	subscriptionRepo := subscription.NewSubscriptionRepository(database)
	planRepo := plan.NewPlanRepository(database)
	paymentRepo := payment.NewPaymentRepository(database)
	invoiceRepo := invoice.NewInvoiceRepository(database)
	accountRepo := account.NewAccountRepository(database)

	consumer := invConsumer.NewInvoiceConsumer(&invConsumer.InvoiceConsumerDeps{
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
		AccountRepository: accountRepo,
	})

	msgs, err := rabbitMq.CreateConsumer(&rabbitmq.ConsumerOptions{
		Queue:   consts.InvoiceQueue,
		AutoAck: false,
	})
	if err != nil {
		log.Fatalf("[consumer] failed to create consumer: %v", err)
	}

	shared.RunConsumerLoop(msgs, consumer.Handle, "invoice")
}
