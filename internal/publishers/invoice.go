package publishers

import (
	"encoding/json"
	"fmt"
	"log"

	stripeGo "github.com/stripe/stripe-go/v84"

	"adv/go-http/internal/consts"
	"adv/go-http/internal/models"
	rabbitmq "adv/go-http/pkg/rabbitMq"
)

type InvoicePublisher struct {
	rabbitMq *rabbitmq.RabbitMq
}

func NewInvoicePublisher(rabbitMq *rabbitmq.RabbitMq) *InvoicePublisher {
	return &InvoicePublisher{
		rabbitMq: rabbitMq,
	}
}

func (p InvoicePublisher) CreateExchangeAndQueue() {
	_, err := p.rabbitMq.CreateExchange(&rabbitmq.ExchangeOptions{
		Name:    consts.InvoiceExchange,
		Type:    "direct",
		Durable: true,
	})
	if err != nil {
		fmt.Print(err)
	}

	_, _, errQ := p.rabbitMq.CreateQueueWithBinding(&rabbitmq.QueueWithBindingOptions{
		QueueOptions: &rabbitmq.QueueOptions{
			Name:       consts.InvoiceQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		Exchange:   consts.InvoiceExchange,
		RoutingKey: consts.InvoiceRouting,
	})
	if errQ != nil {
		fmt.Print(errQ)
	}
}

func (p InvoicePublisher) PublishDataRawToQueue(eventType models.InvoiceEventType, dataRaw *stripeGo.EventData) error {
	var inv stripeGo.Invoice
	if err := json.Unmarshal(dataRaw.Raw, &inv); err != nil {
		return err
	}

	log.Printf("[stripe] %s: id=%s amount=%d\n", eventType, inv.ID, inv.AmountPaid)

	body, err := json.Marshal(models.InvoiceMessage{
		EventType: eventType,
		Data:      inv,
	})
	if err != nil {
		return err
	}

	p.rabbitMq.Publish(&rabbitmq.PublisherOptions{
		Exchange:    consts.InvoiceExchange,
		RoutingKey:  consts.InvoiceRouting,
		Body:        body,
		ContentType: "application/json",
	})

	return nil
}
