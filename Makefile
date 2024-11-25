run: build
	@./bin/ecom

migration:
	@migrate create -ext sql -dir migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

start-dev:
	@go run cmd/main.go

migrate-up:
	@go run migrate/main.go up

migrate-down:
	@go run migrate/main.go down
