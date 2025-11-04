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
	docker compose -f docker-compose.dev.yml exec -it app /bin/bash
.PHONY: docker-exec-app-dev

exec-test: # Build and run test containers, execute tests, and clean up
	docker compose -f docker-compose.test.yml build
	docker compose -f docker-compose.test.yml up -d
	docker compose -f docker-compose.test.yml exec -T app sh run_tests.sh
	docker compose -f docker-compose.test.yml down -v
.PHONY: exec-test
