package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	stripeGo "github.com/stripe/stripe-go/v84"

	"adv/go-http/cmd/shared"
	"adv/go-http/internal/consts"
	piConsumer "adv/go-http/internal/consumers"
	"adv/go-http/internal/payments/invoice"
	"adv/go-http/internal/payments/payment"
	"adv/go-http/internal/payments/plan"
	stripeServices "adv/go-http/internal/payments/stripe/services"
	"adv/go-http/internal/payments/subscription"
	"adv/go-http/pkg/db"
	rabbitmq "adv/go-http/pkg/rabbitMq"
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
			StripeClient:      stripeClient,
			InvoiceRepository: invoiceRepo,
			PaymentRepository: paymentRepo,
		}),
		SubscriptionSvc: subscription.NewSubscriptionService(subscription.SubscriptionServiceDeps{
			SubscriptionRepository: subscriptionRepo,
			PlanRepository:         planRepo,
			StripeClient:           stripeClient,
			Ctx:                    context.Background(),
		}),
	})

	msgs, err := rabbitMq.CreateConsumer(&rabbitmq.ConsumerOptions{
		Queue:   consts.PaymentIntentQueueSucceed,
		AutoAck: false,
	})
	if err != nil {
		log.Fatalf("[consumer] failed to create consumer: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[consumer] waiting for payment_intent messages...")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[consumer] channel closed")
				return
			}
			if err := consumer.Handle(msg.Body); err != nil {
				log.Printf("[consumer] failed to handle message: %v", err)
				var stripeErr *stripeGo.Error
				if errors.As(err, &stripeErr) && stripeErr.HTTPStatusCode == 400 {
					log.Printf("[consumer] stripe 400, discarding message")
					msg.Ack(false)
					continue
				}
				msg.Nack(false, true)
				continue
			}
			msg.Ack(false)
		case <-quit:
			log.Println("[consumer] shutting down...")
			return
		}
	}
}
