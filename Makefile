APP_NAME=commerce-core-go
CMD_DIR=cmd/api
BINARY=bin/app.exe

DOCKER_ENV_FILE=configs/.env.docker
TEST_ENV_FILE=configs/.env.test

DEV_DATABASE_URL ?= postgres://admin:admin@localhost:5432/mydb?sslmode=disable
TEST_DATABASE_URL ?= postgres://test:test@localhost:5433/commerce_test?sslmode=disable

COMPOSE_DEV=docker compose --env-file $(DOCKER_ENV_FILE) -f deployments/docker/docker-compose.dev.yml
COMPOSE_TEST=docker compose --env-file $(TEST_ENV_FILE) -f deployments/docker/docker-compose.test.yml

.PHONY: help \
	run dev build tidy \
	dev-up dev-up-build dev-down dev-down-v dev-logs dev-restart \
	migrate-create migrate-up migrate-down migrate-force migrate-version \
	migrate-test-up migrate-test-down \
	test test-v test-unit test-integration test-integration-v \
	test-up test-up-build test-down test-down-v test-logs lint

help:
	@echo "Available commands:"
	@echo "make dev              - run app locally"
	@echo "make dev-watch        - run app with hot reload"
	@echo "make build            - build app binary"
	@echo "make test             - run all tests"
	@echo "make test-v           - run tests verbose"
	@echo "make test-short       - run short tests"
	@echo "make dev-up           - start docker dev"
	@echo "make dev-up-build     - build and start docker dev"
	@echo "make test-up          - start test database"
	@echo "make migrate-up       - run dev migrations"
	@echo "make migrate-test-up  - run test migrations"

# ========================
# 🟢 Local
# ========================

run:
	go run $(CMD_DIR)/main.go

dev:
	go run $(CMD_DIR)/main.go

dev-watch:
	air

build:
	go build -o $(BINARY) $(CMD_DIR)/main.go

tidy:
	go mod tidy

# ========================
# 🔍 Quality
# ========================

lint:
	golangci-lint run --timeout=5m

# ========================
# 🐳 Docker (dev)
# ========================

dev-up:
	$(COMPOSE_DEV) up

dev-up-build:
	$(COMPOSE_DEV) up --build

dev-down:
	$(COMPOSE_DEV) down

dev-down-remove:
	$(COMPOSE_DEV) down -v

dev-logs:
	$(COMPOSE_DEV) logs -f

dev-restart:
	$(COMPOSE_DEV) down
	$(COMPOSE_DEV) up --build

# ========================
# 🗄️ Migration
# ========================

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "$(DEV_DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DEV_DATABASE_URL)" down 1

migrate-force:
	migrate -path migrations -database "$(DEV_DATABASE_URL)" force $(version)

migrate-version:
	migrate -path migrations -database "$(DEV_DATABASE_URL)" version

migrate-test-up:
	migrate -path migrations -database "$(TEST_DATABASE_URL)" up

migrate-test-down:
	migrate -path migrations -database "$(TEST_DATABASE_URL)" down


# ========================
# 🧪 Test Database
# ========================

test-up:
	$(COMPOSE_TEST) up -d

test-up-build:
	$(COMPOSE_TEST) up --build -d

test-down:
	$(COMPOSE_TEST) down

test-down-remove:
	$(COMPOSE_TEST) down -v

test-logs:
	$(COMPOSE_TEST) logs -f

# ========================
# 🧪 Tests
# ========================

test:
	go test ./... -count=1 -p=1

test-v:
	go test ./... -count=1 -p=1 -v

test-short:
	go test ./... -short -count=1