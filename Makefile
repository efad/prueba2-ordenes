.PHONY: migrate-up migrate-down migrate-status postgres-up postgres-down

DATABASE_URL ?= postgres://orders:orders@localhost:5432/orders?sslmode=disable

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir migrations postgres "$(DATABASE_URL)" status

postgres-up:
	docker compose up postgres -d

postgres-down:
	docker compose down
