package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMq struct {
	conn *amqp.Connection
}

type QueueOptions struct {
	name       string
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
	args       amqp.Table
}

type PublisherOptions struct {
	exchange    string `default:""`
	queue       *amqp.Queue
	channel     *amqp.Channel
	body        []byte
	contentType string `default:"application/json"`
	mandatory   bool   `default:"false"`
	immediate   bool   `default:"false"`
}

type ExchangeOptions struct {
	name       string
	kind       string
	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
	args       amqp.Table
}

type ConsumerOptions struct {
	channel   *amqp.Channel
	queue     string
	consumer  string
	autoAck   bool       `default:"true"`
	exclusive bool       `default:"false"`
	noLocal   bool       `default:"false"`
	noWait    bool       `default:"false"`
	args      amqp.Table `default:"nil"`
}

func NewRabbitMq(rabbutMqUrl string) *RabbitMq {
	conn, err := amqp.Dial(rabbutMqUrl)

	if err != nil {
		fmt.Print(err)
	}

	return &RabbitMq{
		conn: conn,
	}
}

func (r RabbitMq) CreateQueue(queueOptions *QueueOptions) (*amqp.Queue, *amqp.Channel, error) {
	defer r.conn.Close()
	ch, err := r.conn.Channel()

	if err != nil {
		fmt.Print(err)
		return nil, nil, err
	}

	q, queueErr := ch.QueueDeclare(
		queueOptions.name,
		queueOptions.durable,
		queueOptions.autoDelete,
		queueOptions.exclusive,
		queueOptions.noWait,
		queueOptions.args,
	)

	if queueErr != nil {
		fmt.Print(err)
		return nil, nil, err
	}

	return &q, ch, nil
}

func (r RabbitMq) CreateExchange(opts *ExchangeOptions) (*amqp.Channel, error) {
	defer r.conn.Close()
	ch, err := r.conn.Channel()

	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	exchangeErr := ch.ExchangeDeclare(
		opts.name,
		opts.kind,
		opts.durable,
		opts.autoDelete,
		opts.internal,
		opts.noWait,
		opts.args,
	)

	if exchangeErr != nil {
		fmt.Print(exchangeErr)
		return nil, exchangeErr
	}

	return ch, nil
}

func (r RabbitMq) Publish(opt *PublisherOptions) {
	defer r.conn.Close()
	ctx := context.Background()
	err := opt.channel.PublishWithContext(ctx,
		opt.exchange,
		opt.queue.Name,
		opt.mandatory,
		opt.immediate,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  opt.contentType,
			Body:         opt.body,
		},
	)

	if err != nil {
		fmt.Print(err)
		return
	}

}

func (r RabbitMq) CreateConsumer(opts *ConsumerOptions) (<-chan amqp.Delivery, error) {
	defer r.conn.Close()
	ch, err := opts.channel.Consume(
		opts.queue,
		opts.consumer,
		opts.autoAck,
		opts.exclusive,
		opts.noLocal,
		opts.noWait,
		opts.args,
	)

	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	return ch, nil
}
