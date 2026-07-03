include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: migrate-create migrate-up migrate-down seed generate-swag-docs test

migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_DSN) up

migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_DSN) down $(filter-out $@,$(MAKECMDGOALS))

seed:
	@go run cmd/migrate/seed/main.go

generate-swag-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

test:
	@go test -v ./...

stress-test:
	@npx autocannon -r 4000 -d 2 -c 10 --renderStatusCodes http://localhost:8080/v1/health