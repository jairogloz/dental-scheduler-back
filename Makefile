.PHONY: build run test clean deps db-config migrate-up migrate-down migrate-status migrate-force migrate-drop migrate-create migrate-psql docker-build docker-run

# Build the application
build:
	go build -o bin/dental-scheduler ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Database configuration (can be overridden with environment variables)
DB_HOST ?= aws-1-us-east-2.pooler.supabase.com
DB_PORT ?= 5432
DB_USER ?= postgres.wxwpuooiyeblrqntyewq
DB_PASSWORD ?= 5a9MBf1AJlWpHHIN
DB_NAME ?= postgres
DB_SSL_MODE ?= disable

# Construct database URL
DB_URL = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Show current database configuration
db-config:
	@echo "Database Configuration:"
	@echo "  Host: $(DB_HOST)"
	@echo "  Port: $(DB_PORT)"
	@echo "  User: $(DB_USER)"
	@echo "  Password: [HIDDEN]"
	@echo "  Database: $(DB_NAME)"
	@echo "  SSL Mode: $(DB_SSL_MODE)"
	@echo "  Full URL: postgresql://$(DB_USER):[HIDDEN]@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)"

# Database migrations
migrate-up:
	migrate -path internal/infra/database/postgres/migrations -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path internal/infra/database/postgres/migrations -database "$(DB_URL)" -verbose down

migrate-status:
	migrate -path internal/infra/database/postgres/migrations -database "$(DB_URL)" version

migrate-force:
	migrate -path internal/infra/database/postgres/migrations -database "$(DB_URL)" force $(version)

migrate-drop:
	migrate -path internal/infra/database/postgres/migrations -database "$(DB_URL)" drop

migrate-create:
	migrate create -ext sql -dir internal/infra/database/postgres/migrations -seq $(name)

# Alternative: Run migrations using psql
migrate-psql:
	psql "$(DB_URL)" -f internal/infra/database/postgres/migrations/000001_init.up.sql

# Docker commands
docker-build:
	docker build -t dental-scheduler-backend .

docker-run:
	docker-compose up --build

docker-down:
	docker-compose down

# Development helpers
dev:
	air # Requires air for hot reload: go install github.com/cosmtrek/air@latest

fmt:
	go fmt ./...

lint:
	golangci-lint run # Requires golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate mocks for testing
mocks:
	mockgen -source=internal/domain/ports/repositories/appointment_repository.go -destination=internal/domain/ports/repositories/mocks/appointment_repository_mock.go
	mockgen -source=internal/domain/ports/repositories/clinic_repository.go -destination=internal/domain/ports/repositories/mocks/clinic_repository_mock.go
	mockgen -source=internal/domain/ports/repositories/doctor_repository.go -destination=internal/domain/ports/repositories/mocks/doctor_repository_mock.go
	mockgen -source=internal/domain/ports/repositories/patient_repository.go -destination=internal/domain/ports/repositories/mocks/patient_repository_mock.go
	mockgen -source=internal/domain/ports/repositories/unit_repository.go -destination=internal/domain/ports/repositories/mocks/unit_repository_mock.go
