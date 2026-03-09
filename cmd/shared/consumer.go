package shared

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	stripeGo "github.com/stripe/stripe-go/v84"
)

func RunConsumerLoop(msgs <-chan amqp.Delivery, handle func([]byte) error, logPrefix string) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("[%s] waiting for messages...", logPrefix)

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("[%s] channel closed", logPrefix)
				return
			}
			if err := handle(msg.Body); err != nil {
				log.Printf("[%s] failed to handle message: %v", logPrefix, err)
				var stripeErr *stripeGo.Error
				if errors.As(err, &stripeErr) && stripeErr.HTTPStatusCode == 400 {
					log.Printf("[%s] stripe 400, discarding message", logPrefix)
					msg.Ack(false)
					continue
				}
				msg.Nack(false, true)
				continue
			}
			msg.Ack(false)
		case <-quit:
			log.Printf("[%s] shutting down...", logPrefix)
			return
		}
	}
}
