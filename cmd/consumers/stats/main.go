package main

import (
	"log"
	"strconv"
	"time"

	goRedis "github.com/go-redis/redis/v8"

	"link-generator/cmd/shared"
	"link-generator/internal/consts"
	statsConsumer "link-generator/internal/consumers"
	"link-generator/internal/publishers"
	"link-generator/internal/stats"
	pkgClickhouse "link-generator/pkg/clickhouse"
	pkgRedis "link-generator/pkg/redis"
	rabbitmq "link-generator/pkg/rabbitMq"
)

func main() {
	cfg := shared.LoadConfigs()

	rabbitMq := rabbitmq.NewRabbitMq(cfg.RabbitMq)
	defer rabbitMq.Close()

	ch, err := pkgClickhouse.NewCliсkhouse(&cfg.ClickHouse)
	if err != nil {
		log.Fatalf("clickhouse init: %v", err)
	}
	defer ch.Close()

	statsRepo := stats.NewStatsRepository(ch)

	cacheMinutes, _ := strconv.Atoi(cfg.Redis.Cache)
	redisClient := pkgRedis.NewRedis(&goRedis.Options{
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
	}, time.Duration(cacheMinutes)*time.Minute)
	defer redisClient.Close()

	publishers.NewStatsPublisher(rabbitMq).CreateExchangeAndQueue()

	consumer := statsConsumer.NewStatsConsumer(statsRepo, redisClient)

	msgs, err := rabbitMq.CreateConsumer(&rabbitmq.ConsumerOptions{
		Queue:   consts.StatsQueue,
		AutoAck: false,
	})
	if err != nil {
		log.Fatalf("[consumer] failed to create consumer: %v", err)
	}

	shared.RunConsumerLoop(msgs, consumer.Handle, "stats")
}
