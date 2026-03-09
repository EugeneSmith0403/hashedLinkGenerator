package publishers

import (
	"adv/go-http/internal/consts"
	"adv/go-http/internal/models"
	rabbitmq "adv/go-http/pkg/rabbitMq"
	"encoding/json"
	"fmt"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type PaymentPublisher struct {
	rabbitMq *rabbitmq.RabbitMq
}

func NewPaymentPublisher(rabbitMq *rabbitmq.RabbitMq) *PaymentPublisher {
	return &PaymentPublisher{
		rabbitMq: rabbitMq,
	}
}

func (p PaymentPublisher) CreateExchangeAndQueue() {
	_, err := p.rabbitMq.CreateExchange(&rabbitmq.ExchangeOptions{
		Name:    consts.PaymentIntentExchange,
		Type:    "direct",
		Durable: true,
	})

	if err != nil {
		fmt.Print(err)
	}

	_, _, errQ := p.rabbitMq.CreateQueueWithBinding(&rabbitmq.QueueWithBindingOptions{
		QueueOptions: &rabbitmq.QueueOptions{
			Name:       consts.PaymentIntentQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		Exchange:   consts.PaymentIntentExchange,
		RoutingKey: consts.PaymentIntentRouting,
	})

	if errQ != nil {
		fmt.Print(errQ)
	}
}

func (p PaymentPublisher) PublishDataRawToQueue(eventType models.PaymentIntentEventType, dataRaw *stripeGo.EventData) error {
	var pi stripeGo.PaymentIntent
	if err := json.Unmarshal(dataRaw.Raw, &pi); err != nil {
		return err
	}

	log.Printf("[stripe] %s: id=%s amount=%d\n", eventType, pi.ID, pi.Amount)

	body, err := json.Marshal(models.PaymentIntentMessage{
		EventType: models.PaymentIntentEventType(eventType),
		Data:      pi,
	})
	if err != nil {
		return err
	}

	p.rabbitMq.Publish(&rabbitmq.PublisherOptions{
		Exchange:    consts.PaymentIntentExchange,
		RoutingKey:  consts.PaymentIntentRouting,
		Body:        body,
		ContentType: "application/json",
	})

	return nil
}
