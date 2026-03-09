package rabbitmq

import (
	"adv/go-http/configs"
	"context"
	"fmt"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMq struct {
	conn      *amqp.Connection
	Consumers int
}

type QueueOptions struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type QueueWithBindingOptions struct {
	*QueueOptions
	Exchange   string
	RoutingKey string
}

type PublisherOptions struct {
	Exchange    string `default:""`
	RoutingKey  string
	Body        []byte
	ContentType string `default:"application/json"`
	Mandatory   bool   `default:"false"`
	Immediate   bool   `default:"false"`
}

type ExchangeOptions struct {
	Name       string
	Type       string
	Durable    bool       `default:"true"`
	AutoDelete bool       `default:"false"`
	Internal   bool       `default:"false"`
	NoWait     bool       `default:"false"`
	Args       amqp.Table `default:"nil"`
}

type ConsumerOptions struct {
	Queue     string
	Consumer  string
	AutoAck   bool       `default:"true"`
	Exclusive bool       `default:"false"`
	NoLocal   bool       `default:"false"`
	NoWait    bool       `default:"false"`
	Args      amqp.Table `default:"nil"`
}

func NewRabbitMq(rabbitMq configs.RabbitMq) *RabbitMq {
	conn, err := amqp.Dial(rabbitMq.Amqp)

	if err != nil {
		fmt.Print(err)
	}

	consumers, errW := strconv.Atoi(rabbitMq.Consumers)

	if errW != nil {
		fmt.Print(errW)
	}

	return &RabbitMq{
		conn:      conn,
		Consumers: consumers,
	}
}

func (r RabbitMq) Close() {
	r.conn.Close()
}

func (r RabbitMq) CreateQueue(queueOptions *QueueOptions) (*amqp.Queue, *amqp.Channel, error) {
	ch, err := r.conn.Channel()

	if err != nil {
		fmt.Print(err)
		return nil, nil, err
	}

	q, queueErr := ch.QueueDeclare(
		queueOptions.Name,
		queueOptions.Durable,
		queueOptions.AutoDelete,
		queueOptions.Exclusive,
		queueOptions.NoWait,
		queueOptions.Args,
	)

	if queueErr != nil {
		fmt.Print(err)
		return nil, nil, err
	}

	return &q, ch, nil
}

func (r RabbitMq) CreateQueueWithBinding(opts *QueueWithBindingOptions) (*amqp.Queue, *amqp.Channel, error) {
	q, ch, err := r.CreateQueue(opts.QueueOptions)

	if err != nil {
		return nil, nil, err
	}

	bindErr := ch.QueueBind(
		q.Name,
		opts.RoutingKey,
		opts.Exchange,
		opts.NoWait,
		opts.QueueOptions.Args,
	)

	if bindErr != nil {
		return nil, nil, bindErr
	}

	return q, ch, nil

}

func (r RabbitMq) CreateExchange(opts *ExchangeOptions) (*amqp.Channel, error) {
	ch, err := r.conn.Channel()

	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	exchangeErr := ch.ExchangeDeclare(
		opts.Name,
		opts.Type,
		opts.Durable,
		opts.AutoDelete,
		opts.Internal,
		opts.NoWait,
		opts.Args,
	)

	if exchangeErr != nil {
		fmt.Print(exchangeErr)
		return nil, exchangeErr
	}

	return ch, nil
}

func (r RabbitMq) Publish(opt *PublisherOptions) {
	ctx := context.Background()
	ch, _ := r.conn.Channel()

	err := ch.PublishWithContext(ctx,
		opt.Exchange,
		opt.RoutingKey,
		opt.Mandatory,
		opt.Immediate,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  opt.ContentType,
			Body:         opt.Body,
		},
	)

	if err != nil {
		fmt.Print(err)
		return
	}

}

func (r RabbitMq) BindQueue(ch *amqp.Channel, queueName, routingKey, exchange string) error {
	err := ch.QueueBind(queueName, routingKey, exchange, false, nil)
	if err != nil {
		fmt.Print(err)
	}
	return err
}

func (r RabbitMq) CreateConsumer(opts *ConsumerOptions) (<-chan amqp.Delivery, error) {
	ch, _ := r.conn.Channel()
	chC, err := ch.Consume(
		opts.Queue,
		opts.Consumer,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)

	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	return chC, nil
}
