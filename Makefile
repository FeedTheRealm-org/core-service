README = README.md
README_TMP = README.tmp

help: # Show this help message
	@awk -F'#' '/^[^[:space:]].*:/ && !/^\.PHONY/ { \
		target = $$1; \
		comment = ($$2 ? $$2 : ""); \
		printf "  %-48s %s\n", target, comment \
	}' Makefile
.PHONY: help

update-tree-structure: # Update the project tree structure file
	@awk '/## Structure/{exit} {print}' $(README) > $(README_TMP)
	@cat $(README_TMP) > $(README)
	@rm $(README_TMP)
	@printf '\n## Structure\n\n' >> $(README)
	@printf 'Se tiene la siguiente estructura base, donde cada microservicio que forma parte del monolith se separa del resto para eventualmente escalar la arquitectura,\n' >> $(README)
	@printf 'en cada uno los controladores, servicios y repositorios se ponen en sus correspontientes carpetas.\n\n' >> $(README)
	@printf '%s\n' '- **Crear su archivo separado para la interfaz y otro para la implementacion**.' >> $(README)
	@printf '%s\n\n' '- **No utilizar dependencias de un servicio en otro (no cross-imports)**.' >> $(README)
	@echo '```bash' >> $(README)
	@tree -d --noreport >> $(README)
	@echo '```' >> $(README)
.PHONY: update-tree-structure

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
	docker compose -f docker-compose.dev.yml exec -T app go run ./cmd/migrate/main.go up
	docker compose -f docker-compose.dev.yml exec -it app /bin/bash
.PHONY: docker-exec-app-dev

exec-test: # Build and run test containers, execute tests, and clean up
	docker compose -f docker-compose.test.yml build
	docker compose -f docker-compose.test.yml up -d
	docker compose -f docker-compose.test.yml exec -T app sh run_tests.sh
	docker compose -f docker-compose.test.yml down -v
.PHONY: exec-test
