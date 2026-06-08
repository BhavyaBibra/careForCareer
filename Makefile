.PHONY: build run-api run-worker test lint migrate-up migrate-down seed docker-up docker-down keys tidy

BINARY_API    := bin/api
BINARY_WORKER := bin/worker

# Load .env if present so DATABASE_URL etc. are available
ifneq (,$(wildcard .env))
  include .env
  export $(shell sed 's/=.*//' .env | grep -v '^\#')
endif

MIGRATE := migrate -path ./migrations -database "$(DATABASE_URL)"
DOCKER_COMPOSE := $(shell docker compose version > /dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

build:
	go build -o $(BINARY_API) ./cmd/api
	go build -o $(BINARY_WORKER) ./cmd/worker

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

seed:
	go run ./seed/main.go

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

keys:
	mkdir -p config/keys
	openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out config/keys/private.pem
	openssl pkey -in config/keys/private.pem -pubout -out config/keys/public.pem
	@echo "Keys generated at config/keys/"

tidy:
	go mod tidy
