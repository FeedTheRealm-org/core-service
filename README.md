# Feed the Realm Core-service

[![codecov](https://codecov.io/gh/FeedTheRealm-org/core-service/graph/badge.svg?token=S7NI26WBRL)](https://codecov.io/gh/FeedTheRealm-org/core-service)

Monolithic backend for all adjacent services to the game.

**Structure must always stay easy to divide into microservices and escalate if needed.**

## Documentation

For detailed information, please refer to the specific documentation files in the `docs/` directory:

- [Architecture & Modular Monolith Details](docs/architecture.md)
- [Development Guide & Local Setup](docs/development.md)
- [Production Usage & Deployment](docs/production.md)
- [Testing Setup](docs/testing.md)
- [CI/CD & GitHub Actions](docs/github_actions.md)

## Technologies

- [Gin-gonic](https://gin-gonic.com/)
- [Go](https://go.dev/)
- [Docker](https://docs.docker.com/)
- [Swaggo](https://github.com/swaggo/swag)
- [Nomad (via hashicorp)](https://developer.hashicorp.com/nomad)
- [golang-migrate](https://github.com/golang-migrate/migrate)

## Quick Start

### Dependencies

- Install [golang](https://go.dev/doc/install)
- Install [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- Install `make`

### Development

```bash
# Setup environment variables
cp .env.example .env

# Run development environment (API, DB, Buckets)
make dev
```

### Production

How to use docker to build and run server.

```bash
# Setup environment variables
cp .env.example .env

# Run production environment
make up-build
```

## How to test

See the [Testing Guide](docs/testing.md) for more details.

```bash
# Run tests in Docker (includes acceptance tests)
make test
```

## Makefile Commands

The project includes a `Makefile` with convenient commands for development and testing. See [Development Guide](docs/development.md) for more usage examples.

```bash
# Show all available commands with descriptions
make help

# Development commands
make dev          # Starts detached containers and an interactive shell
make up-build     # Builds & starts production profile containers, or just: make up
make build        # Builds production profile containers
make down         # Stops all running containers
make logs         # Tail logs from all containers
make logs-<svc>   # Tail logs from a specific service (e.g. make logs-db)
make db           # Open a psql shell in the Postgres container
make clean        # Remove all containers, images, and local bucket data

# Testing commands
make test             # Build, run, and execute tests in a clean Docker environment
make test-unit        # Run Go unit tests only
make test-acceptance  # Run Python/behave acceptance tests only

# Database migrations
make migration service=<name> name=<migration_name>  # Create a new migration file

# Data seeding
make seed ASSETS_BASE_PATH=<path>                          # Seed local environment
make seed-prod ASSETS_BASE_PATH=<path> JWT_TOKEN=<token>   # Seed production

# Documentation
make swagger      # Generate Swagger documentation
```

## Database Migrations

The project uses `golang-migrate` for database migrations. Migration files are located in the `migrations/` directory. For setup details, see [Development Guide](docs/development.md).

## Structure

See [Architecture Documentation](docs/architecture.md) for a comprehensive breakdown of the modular monolith design.

## Swagger Documentation

Endpoint documentation was made with `Swagger`. Once starting up the project, navigate to this link to test out the endpoints:

```sh
http://localhost:8000/swagger/index.html
```

## Maintenance Scripts

Check the [Development Guide](docs/development.md) for full descriptions of maintenance scripts like `seed_initial_cosmetics.py`.
