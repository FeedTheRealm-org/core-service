README = README.md
README_TMP = README.tmp

help: # Show this help message
	@awk -F'#' '/^[^[:space:]].*:/ && !/^\.PHONY/ { \
		target = $$1; \
		comment = ($$2 ? $$2 : ""); \
		printf "  %-48s %s\n", target, comment \
	}' Makefile
.PHONY: help

down: # Stop and remove containers
	docker compose -f docker-compose.yml down
.PHONY: down

up: down # Build and start containers
	docker compose -f docker-compose.yml build
	docker compose -f docker-compose.yml up -d
.PHONY: up

docker-down-dev: # Stop and remove development containers
	docker compose -f docker-compose.dev.yml down
.PHONY: docker-down-dev

docker-build-dev: docker-down-dev # Build development containers
	docker compose -f docker-compose.dev.yml build
.PHONY: docker-build-dev

docker-up-dev: docker-build-dev # Start development containers
	docker compose -f docker-compose.dev.yml up -d
.PHONY: docker-up-dev

docker-exec-app-dev: docker-up-dev # Execute a bash shell in the development app container
	docker compose -f docker-compose.dev.yml exec app swag init -g cmd/main.go -o ./swagger
	docker compose -f docker-compose.dev.yml exec -it app /bin/bash
.PHONY: docker-exec-app-dev

docker-run-app-dev: docker-up-dev # Run the application in the development container
	docker compose -f docker-compose.dev.yml exec app swag init -g cmd/main.go -o ./swagger
	docker compose -f docker-compose.dev.yml exec app go run cmd/main.go
.PHONY: docker-run-app-dev

exec-test:
	docker compose -f docker-compose.test.yml down -v --remove-orphans
	docker compose -f docker-compose.test.yml build
	docker compose -f docker-compose.test.yml up -d --remove-orphans
	docker compose -f docker-compose.test.yml exec -T app sh run_tests.sh
	docker compose -f docker-compose.test.yml down -v --remove-orphans
.PHONY: exec-test



swag init:
	swag init -g cmd/main.go -o ./swagger
.PHONY: swag init


migrate-create:
	migrate create -ext sql -dir migrations $(name)
.PHONY: migrate-create
