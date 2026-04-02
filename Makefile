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
	docker compose -f $(COMPOSE_DEV) down --remove-orphans -t 2
	docker network prune -f
.PHONY: down

build: down # Build containers
	docker compose -f $(COMPOSE_DEV) --profile prod build
.PHONY: build

up: down # Start containers
	docker compose -f $(COMPOSE_DEV) --profile prod up --force-recreate
.PHONY: up

up-build: down # Build and start containers
	docker compose -f $(COMPOSE_DEV) --profile prod up --build --force-recreate
.PHONY: up-build

dev: # Starts containers detatched and starts interactive shell in app container for manual runs
	docker compose -f $(COMPOSE_DEV) --profile dev up -d --build app-dev db buckets stripe-webhook-dev
	docker compose -f $(COMPOSE_DEV) exec app-dev swag init -g cmd/main.go -o ./swagger
	docker compose -f $(COMPOSE_DEV) exec app-dev /bin/bash
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
	docker compose -f $(COMPOSE_DEV) down --remove-orphans -v
	docker compose -f $(COMPOSE_TEST) down --remove-orphans -v
.PHONY: clean

swagger: # Generate Swagger documentation
	swag init -g cmd/main.go -o ./swagger
.PHONY: swagger

migration: # Create a new database migration. Usage: make migration service=your_service name=your_migration_name
	migrate create -ext sql -dir migrations/$(service) $(name)
.PHONY: migration

logs: # Tail logs of all containers
	docker compose -f $(COMPOSE_DEV) logs -f
.PHONY: logs

logs-%: # Tail logs of a specific service. Usage: make logs-service_name
	docker compose -f $(COMPOSE_DEV) logs -f $*
.PHONY: logs-%
