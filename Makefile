.PHONY: dev server consumers consumer-payment consumer-subscription consumer-invoice frontend build docker-up docker-down migrate

# Read from .env; override via CLI: make consumers WORKERS=3
WORKERS ?= $(shell grep -E '^RABBITMQ_CONSUMERS' .env | cut -d'=' -f2 | tr -d '"' 2>/dev/null || echo 1)

dev: docker-up migrate
	@trap 'kill 0' INT TERM; \
	go run ./cmd/server & \
	until nc -z localhost 8081 2>/dev/null; do sleep 0.2; done; \
	for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/paymentIntent & \
		go run ./cmd/consumers/subscription & \
		go run ./cmd/consumers/invoice & \
	done; \
	cd frontend && pnpm install && pnpm dev & \
	wait

server:
	go run ./cmd/server

consumers:
	@for i in $$(seq 1 $(WORKERS)); do \
		go run ./cmd/consumers/paymentIntent & \
		go run ./cmd/consumers/subscription & \
		go run ./cmd/consumers/invoice & \
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

frontend:
	cd frontend && pnpm install && pnpm dev

docker-up:
	docker compose up -d --wait

docker-down:
	docker compose down

migrate:
	go run ./migrations

build:
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/consumer-payment ./cmd/consumers/paymentIntent
	go build -o bin/consumer-subscription ./cmd/consumers/subscription
	go build -o bin/consumer-invoice ./cmd/consumers/invoice
