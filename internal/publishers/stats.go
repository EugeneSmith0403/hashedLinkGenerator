package publishers

import (
	"encoding/json"
	"fmt"
	"log"

	"link-generator/internal/consts"
	"link-generator/internal/models"
	rabbitmq "link-generator/pkg/rabbitMq"
)

type StatsPublisher struct {
	rabbitMq *rabbitmq.RabbitMq
}

func NewStatsPublisher(rabbitMq *rabbitmq.RabbitMq) *StatsPublisher {
	return &StatsPublisher{
		rabbitMq: rabbitMq,
	}
}

func (p StatsPublisher) CreateExchangeAndQueue() {
	_, err := p.rabbitMq.CreateExchange(&rabbitmq.ExchangeOptions{
		Name:    consts.StatsExchange,
		Type:    "direct",
		Durable: true,
	})
	if err != nil {
		fmt.Print(err)
	}

	_, _, errQ := p.rabbitMq.CreateQueueWithBinding(&rabbitmq.QueueWithBindingOptions{
		QueueOptions: &rabbitmq.QueueOptions{
			Name:       consts.StatsQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		Exchange:   consts.StatsExchange,
		RoutingKey: consts.StatsRouting,
	})
	if errQ != nil {
		fmt.Print(errQ)
	}
}

func (p StatsPublisher) PublishToQueue(eventType models.StatsEventType, data models.LinkTransition) error {
	log.Printf("[stats] %s: LinkID=%s ClickedAt=%d\n", eventType, data.LinkID, data.ClickedAt)

	body, err := json.Marshal(models.StatsMessage{
		EventType: eventType,
		Data:      data,
	})
	if err != nil {
		return err
	}

	p.rabbitMq.Publish(&rabbitmq.PublisherOptions{
		Exchange:    consts.StatsExchange,
		RoutingKey:  consts.StatsRouting,
		Body:        body,
		ContentType: "application/json",
	})

	return nil
}
