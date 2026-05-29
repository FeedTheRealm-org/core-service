# Development Guide

This guide covers everything you need to know to run the core-service locally.

## Prerequisites

- [Go](https://go.dev/doc/install)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

## Useful Links

- [Gin-gonic Documentation](https://gin-gonic.com/docs/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Swaggo](https://github.com/swaggo/swag)

## Setup Steps

1. **Environment Variables:**
   Copy the example environment file and configure it.

    ```bash
    cp .env.example .env
    ```

2. **Start Development Environment:**
   Using the provided Makefile, you can spin up the application, database, and bucket storage locally.

    ```bash
    make dev
    ```

    _Note: This starts the containers detached and provides an interactive shell for manual executions if needed._

3. **Running Locally (without Docker):**
    ```bash
    go run cmd/main.go
    ```

## Database Migrations

Migrations are handled via `golang-migrate`.

- Automatic execution is handled on server start if `DB_SHOULD_MIGRATE=true`.
- To generate a new migration:
    ```bash
    make migration service=<service_name> name=<migration_name>
    ```

## API Documentation (Swagger)

Swagger docs are automatically generated.

- Generate docs locally: `make swagger` or `swag init -g cmd/main.go -o ./swagger`
- Access Swagger UI: `http://localhost:8000/swagger/index.html`

## Useful Scripts

Scripts live under the `scripts/` directory and are invoked via `make seed` or `make seed-prod`.

- `scripts/seed_initial_cosmetics.py`: Seeds cosmetic assets.
- `scripts/seed_default_models.py`: Seeds default 3D models.
- `scripts/seed_default_materials.py`: Seeds default materials.
- `scripts/seed_bot_accounts.py`: Seeds bot player accounts.
