package publishers

import (
	"link-generator/internal/consts"
	"link-generator/internal/models"
	rabbitmq "link-generator/pkg/rabbitMq"
	"encoding/json"
	"fmt"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"
)

type SubscriptionPublisher struct {
	rabbitMq *rabbitmq.RabbitMq
}

func NewSubscriptionPublisher(rabbitMq *rabbitmq.RabbitMq) *SubscriptionPublisher {
	return &SubscriptionPublisher{
		rabbitMq: rabbitMq,
	}
}

func (p SubscriptionPublisher) CreateExchangeAndQueue() {
	_, err := p.rabbitMq.CreateExchange(&rabbitmq.ExchangeOptions{
		Name:    consts.SubscriptionExchange,
		Type:    "direct",
		Durable: true,
	})

	if err != nil {
		fmt.Print(err)
	}

	_, _, errQ := p.rabbitMq.CreateQueueWithBinding(&rabbitmq.QueueWithBindingOptions{
		QueueOptions: &rabbitmq.QueueOptions{
			Name:       consts.SubscriptionQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		Exchange:   consts.SubscriptionExchange,
		RoutingKey: consts.SubscriptionRouting,
	})

	if errQ != nil {
		fmt.Print(errQ)
	}
}

func (p SubscriptionPublisher) PublishDataRawToQueue(eventType models.SubscriptionEventType, dataRaw *stripeGo.EventData) error {
	var sub stripeGo.Subscription
	if err := json.Unmarshal(dataRaw.Raw, &sub); err != nil {
		return err
	}

	log.Printf("[stripe] %s: id=%s\n", eventType, sub.ID)

	body, err := json.Marshal(models.SubscriptionMessage{
		EventType: models.SubscriptionEventType(eventType),
		Data:      sub,
	})
	if err != nil {
		return err
	}

	p.rabbitMq.Publish(&rabbitmq.PublisherOptions{
		Exchange:    consts.SubscriptionExchange,
		RoutingKey:  consts.SubscriptionRouting,
		Body:        body,
		ContentType: "application/json",
	})

	return nil
}
