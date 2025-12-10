COMPOSE_BASE := docker-compose.yml
COMPOSE_DEV := docker-compose.dev.yml
COMPOSE_TEST := docker-compose.test.yml

help: # Show this help message
	@awk -F'#' '/^[^[:space:]].*:/ && !/^\.PHONY/ { \
		target = $$1; \
		comment = ($$2 ? $$2 : ""); \
		printf "  %-48s %s\n", target, comment \
	}' Makefile
.PHONY: help

down: # Stop and remove containers
	docker compose -f $(COMPOSE_BASE) down
.PHONY: down

up: down # Build and start containers
	docker compose -f $(COMPOSE_BASE) build
	docker compose -f $(COMPOSE_BASE) up -d
.PHONY: up

down-dev: # Stop and remove development containers
	docker compose -f $(COMPOSE_DEV) down
.PHONY: down-dev

build-dev: down-dev # Build development containers
	docker compose -f $(COMPOSE_DEV) build
.PHONY: build-dev

up-dev: build-dev # Start development containers
	docker compose -f $(COMPOSE_DEV) up -d
.PHONY: up-dev

exec-dev: up-dev # Execute a bash shell in the development app container
	docker compose -f $(COMPOSE_DEV) exec app swag init -g cmd/main.go -o ./swagger
	-docker compose -f $(COMPOSE_DEV) exec -it app /bin/bash
	docker compose -f $(COMPOSE_DEV) down
.PHONY: exec-dev

run-dev: up-dev # Run the application in the development container
	docker compose -f $(COMPOSE_DEV) exec app swag init -g cmd/main.go -o ./swagger
	docker compose -f $(COMPOSE_DEV) exec app go run cmd/main.go
	docker compose -f $(COMPOSE_DEV) down -v --remove-orphans
.PHONY: run-dev

exec-test:
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
	docker compose -f $(COMPOSE_TEST) build
	docker compose -f $(COMPOSE_TEST) up -d --remove-orphans
	docker compose -f $(COMPOSE_TEST) exec -T app sh run_tests.sh
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
.PHONY: exec-test

swag init:
	swag init -g cmd/main.go -o ./swagger
.PHONY: swag init

migrate-create:
	migrate create -ext sql -dir migrations $(name)
.PHONY: migrate-create
