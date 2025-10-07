

docker-down-dev:
	docker compose -f docker-compose.dev.yml down
.PHONY: docker-down-dev

docker-build-dev: docker-down-dev
	docker compose -f docker-compose.dev.yml build
.PHONY: docker-build-dev

docker-up-dev: docker-build-dev
	docker compose -f docker-compose.dev.yml up -d
.PHONY: docker-up-dev

docker-exec-app-dev: docker-up-dev
	docker compose -f docker-compose.dev.yml exec -it app /bin/bash
.PHONY: docker-exec-app-dev
