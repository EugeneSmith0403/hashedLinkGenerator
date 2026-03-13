.PHONY: dev server consumers consumer-payment consumer-subscription consumer-invoice consumer-stats frontend build docker-up docker-down migrate migrate-postgres migrate-clickhouse stripe-webhook

# Read from .env; override via CLI: make consumers WORKERS=3
WORKERS ?= $(shell grep -E '^RABBITMQ_CONSUMERS' .env | cut -d'=' -f2 | tr -d '"' 2>/dev/null || echo 1)
STRIPE ?= 0

dev: docker-up migrate
	@trap 'kill 0' INT TERM; \
	go run ./cmd/server & \
	until nc -z localhost 8081 2>/dev/null; do sleep 0.2; done; \
	for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/paymentIntent & \
		go run ./cmd/consumers/subscription & \
		go run ./cmd/consumers/invoice & \
		go run ./cmd/consumers/stats & \
	done; \
	cd frontend && pnpm install && pnpm dev & \
	if [ "$(STRIPE)" = "1" ]; then stripe listen --forward-to localhost:3000/stripe/webhook & fi; \
	wait

server:
	go run ./cmd/server

consumers:
	@for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/paymentIntent & \
		go run ./cmd/consumers/subscription & \
		go run ./cmd/consumers/invoice & \
		go run ./cmd/consumers/stats & \
	done; \
	wait

consumer-payment:
	@for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/paymentIntent & \
	done; \
	wait

consumer-subscription:
	@for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/subscription & \
	done; \
	wait

consumer-invoice:
	@for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/invoice & \
	done; \
	wait

consumer-stats:
	@for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/stats & \
	done; \
	wait

frontend:
	cd frontend && pnpm install && pnpm dev

docker-up:
	docker compose up -d --wait

docker-down:
	docker compose down

migrate:
	go run ./migrations

migrate-postgres:
	go run ./migrations -target=postgres

migrate-clickhouse:
	go run ./migrations -target=clickhouse

stripe-webhook:
	stripe listen --forward-to localhost:3000/stripe/webhook

build:
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/consumer-payment ./cmd/consumers/paymentIntent
	go build -o bin/consumer-subscription ./cmd/consumers/subscription
	go build -o bin/consumer-invoice ./cmd/consumers/invoice
	go build -o bin/consumer-stats ./cmd/consumers/stats
