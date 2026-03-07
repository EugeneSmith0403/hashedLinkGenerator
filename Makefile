.PHONY: dev server consumers frontend build docker-up docker-down

# Read from .env; override via CLI: make consumers CONSUMERS=3
CONSUMERS ?= $(shell grep -E '^RABBITMQ_CONSUMERS' .env | cut -d'=' -f2 | tr -d '"' 2>/dev/null || echo 1)

dev: docker-up
	@trap 'kill 0' INT TERM; \
	go run ./cmd/server & \
	for i in $$(seq 1 $(CONSUMERS)); do \
		go run ./cmd/consumers & \
	done; \
	cd frontend && pnpm dev & \
	wait

server:
	go run ./cmd/server

consumers:
	@for i in $$(seq 1 $(CONSUMERS)); do \
		go run ./cmd/consumers & \
	done; \
	wait

frontend:
	cd frontend && pnpm dev

docker-up:
	docker compose up -d --wait

docker-down:
	docker compose down

build:
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/consumer ./cmd/consumers
