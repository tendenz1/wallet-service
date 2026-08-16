.PHONY: up down test test-race test-integration test-integration-race vet

TEST_DATABASE_URL ?= postgres://wallet:wallet@localhost:15432/wallet?sslmode=disable

up:
	docker compose up --build

down:
	docker compose down -v

test:
	go test ./...

test-race:
	go test -race ./...

test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -tags=integration ./internal/repository

test-integration-race:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -race -tags=integration ./internal/repository

vet:
	go vet ./...
