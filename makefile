APP_NAME=kart-challenge

.PHONY: up down build run

up:
	docker-compose up -d
	docker-compose ps

down:
	docker-compose down

build:
	go build -o $(APP_NAME) ./cmd/http

migration:
	go run cmd/migration/migration.go

run:
	./$(APP_NAME)

run-all: up build run
