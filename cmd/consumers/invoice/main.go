package main

import (
	"context"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"

	"link-generator/cmd/shared"
	"link-generator/internal/account"
	"link-generator/internal/consts"
	invConsumer "link-generator/internal/consumers"
	"link-generator/internal/locales"
	"link-generator/internal/mailer"
	"link-generator/internal/payments/invoice"
	"link-generator/internal/payments/payment"
	"link-generator/internal/payments/plan"
	"link-generator/internal/payments/subscription"
	"link-generator/internal/publishers"
	"link-generator/pkg/db"
	rabbitmq "link-generator/pkg/rabbitMq"
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

	m := mailer.NewMailer(mailer.MailerDeps{
		LocalesFS:  locales.FS,
		LocalesDir: "invoice/succeed",
		Host:       cfg.Mailer.Host,
		Port:       cfg.Mailer.Port,
		User:       cfg.Mailer.User,
		Password:   cfg.Mailer.Password,
	})

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
		PlanRepository:    planRepo,
		Mailer:            m,
		MailerFrom:        cfg.Mailer.From,
		AppName:           "Go Adv",
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
