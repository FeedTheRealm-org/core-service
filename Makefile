COMPOSE_DEV := docker-compose.dev.yml
COMPOSE_TEST := docker-compose.test.yml

EXEC_APP := go run cmd/main.go
LOCAL_SERVER := http://localhost:8000
PROD_SERVER := https://core.feedtherealm.world
SEED_COSMETICS_SCRIPT := ./scripts/seed_initial_cosmetics.py
SEED_BOTS_SCRIPT := ./scripts/seed_bot_accounts.py

help: # Show this help message
	@awk -F'#' '/^[^[:space:]].*:/ && !/^\.PHONY/ { \
		target = $$1; \
		comment = ($$2 ? $$2 : ""); \
		printf "  %-48s %s\n", target, comment \
	}' Makefile
.PHONY: help

down: # Stop and remove containers
	docker compose -f $(COMPOSE_DEV) --profile prod down --remove-orphans -t 2
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
	docker compose -f $(COMPOSE_DEV) --profile dev up -d --build
	docker compose -f $(COMPOSE_DEV) exec app-dev swag init -g cmd/main.go -o ./swagger
	docker compose -f $(COMPOSE_DEV) exec app-dev /bin/bash
	docker compose -f $(COMPOSE_DEV) --profile dev down
.PHONY: dev

test: # Execute all tests (unit + acceptance)
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
	docker compose -f $(COMPOSE_TEST) --profile acceptance build
	docker compose -f $(COMPOSE_TEST) up -d --wait --remove-orphans
	docker compose -f $(COMPOSE_TEST) --profile acceptance run --rm python-tests
	docker compose -f $(COMPOSE_TEST) run --rm app sh run_tests.sh
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
.PHONY: test

test-unit: # Execute only Go unit tests
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
	docker compose -f $(COMPOSE_TEST) build
	docker compose -f $(COMPOSE_TEST) up test_db -d --wait --remove-orphans
	docker compose -f $(COMPOSE_TEST) run --rm app sh run_tests.sh
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
.PHONY: test-unit

test-acceptance: # Execute only Python/behave acceptance tests
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
	docker compose -f $(COMPOSE_TEST) --profile acceptance build
	docker compose -f $(COMPOSE_TEST) up -d --wait --remove-orphans
	docker compose -f $(COMPOSE_TEST) --profile acceptance run --rm python-tests
	docker compose -f $(COMPOSE_TEST) down -v --remove-orphans
.PHONY: test-acceptance

clean: # Remove all containers and images
	docker compose -f $(COMPOSE_DEV) down --remove-orphans -v
	docker compose -f $(COMPOSE_TEST) down --remove-orphans -v
	sudo rm -rf local_buckets/
.PHONY: clean

seed: # Seed the core-service local resources
ifndef SPRITE_BASE_PATH
	$(error SPRITE_BASE_PATH is required. Usage: SPRITE_BASE_PATH=xxx make seed)
endif
	mkdir -p local_buckets
	chmod -R 777 local_buckets
	docker compose -f $(COMPOSE_DEV) down -v --remove-orphans
	docker compose -f $(COMPOSE_DEV) --profile prod up --build -d --wait --remove-orphans
	export JWT_TOKEN=$$(curl -X POST $(LOCAL_SERVER)/auth/login -H "Content-Type: text/json" -d '{"email": "admin@admin.admin", "password": "admin123"}'  | jq -r '.data.access_token'); \
	$(SEED_COSMETICS_SCRIPT) $(LOCAL_SERVER) $(SPRITE_BASE_PATH) && \
	$(SEED_BOTS_SCRIPT) $(LOCAL_SERVER)
	$(MAKE) down
.PHONY: seed

seed-prod: # Seed the core-service local resources
ifndef SPRITE_BASE_PATH
	$(error SPRITE_BASE_PATH is required. Usage: export SPRITE_BASE_PATH=xxx, then make seed-prod)
endif
ifndef JWT_TOKEN
	$(error $(JWT_HELP))
endif
	$(SEED_COSMETICS_SCRIPT) $(PROD_SERVER) $(SPRITE_BASE_PATH)
	$(SEED_BOTS_SCRIPT) $(PROD_SERVER)
.PHONY: seed

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

db: # Open a psql shell in the postgres container
	docker compose -f $(COMPOSE_DEV) exec db psql -U postgres
.PHONY: db

# --- Messages ---

define JWT_HELP
JWT_TOKEN is required, export it and then run the command again.

Example:

export JWT_TOKEN=$$(curl -X POST $(PROD_SERVER)/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@admin.admin","password":"admin123"}' \
  | jq -r '.data.access_token')
make <command>
endef
