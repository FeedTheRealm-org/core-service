COMPOSE_DEV := docker-compose.dev.yml
COMPOSE_TEST := docker-compose.test.yml

EXEC_APP := go run cmd/main.go

help: # Show this help message
	@awk -F'#' '/^[^[:space:]].*:/ && !/^\.PHONY/ { \
		target = $$1; \
		comment = ($$2 ? $$2 : ""); \
		printf "  %-48s %s\n", target, comment \
	}' Makefile
.PHONY: help

down: # Stop and remove containers
	docker compose -f $(COMPOSE_DEV) down -t 2 --remove-orphans
.PHONY: down

build: down # Build containers
	docker compose -f $(COMPOSE_DEV) --profile prod build
.PHONY: build

up: down # Start containers
	docker compose -f $(COMPOSE_DEV) --profile prod up
.PHONY: up

up-build: down # Build and start containers
	docker compose -f $(COMPOSE_DEV) --profile prod up --build
.PHONY: up-build

dev: # Starts containers detatched and starts interactive shell in app container for manual runs
	docker compose -f $(COMPOSE_DEV) --profile dev up -d --build dev_db buckets
	docker compose -f $(COMPOSE_DEV) run --rm --service-ports app-dev swag init -g cmd/main.go -o ./swagger
	docker compose -f $(COMPOSE_DEV) run --rm --service-ports app-dev /bin/bash
	docker compose -f $(COMPOSE_DEV) down
.PHONY: dev

test: # Execute all tests
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
	docker compose -f $(COMPOSE_TEST) build
	docker compose -f $(COMPOSE_TEST) up -d --remove-orphans
	docker compose -f $(COMPOSE_TEST) exec -T app-dev sh run_tests.sh
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
.PHONY: test

clean: # Remove all containers and images
	docker compose -f $(COMPOSE_BASE) down -v --remove-orphans
	docker compose -f $(COMPOSE_DEV) down -v --remove-orphans
.PHONY: clean

swagger: # Generate Swagger documentation
	swag init -g cmd/main.go -o ./swagger
.PHONY: swagger

migration: # Create a new database migration. Usage: make migration service=your_service name=your_migration_name
	migrate create -ext sql -dir migrations/$(service) $(name)
.PHONY: migration
