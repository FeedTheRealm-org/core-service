# Feed the Realm Core-service

Monolithic backend for all adjacent services to the game.

**Structure must always stay easy to divide into microservices and escalate if needed.**

## Technologies

- Gin-gonic
- Go

## How to run

### Dependencies

- Install golang!
- Intall docker!

### Development

```bash
# Run development instance
go run cmd/main

# Build binary executable
go build cmd/main

# Format code
go fmt ./...
```

### Production

How to use docker to build and run server.

```bash
# Build docker image
docker build -t core-service .

# Run dockerized container
docker run --rm -p <any_port>:8080 core-service:latest

# Cleanup image
docker rmi join-travel-back:latest
```

## How to test

Tests should be written in the same package of the tested object, and the package name should be the same but appending `_test`.

```bash
# Run tests in Docker (includes acceptance tests)
make exec-test
```

## Makefile Commands

The project includes a Makefile with convenient commands for development and testing:

```bash
# Show all available commands with descriptions
make help

# Development commands
make docker-up-dev          # Build and start development containers
make docker-build-dev       # Build development containers
make docker-down-dev        # Stop and remove development containers
make docker-exec-app-dev    # Run migrations and open bash shell in app container

# Testing commands
make exec-test              # Build, run, and execute tests in Docker containers
```

## Database Migrations

The project uses `golang-migrate` for database migrations. Migration files are located in the `migrations/` directory.

```bash
# Run migrations manually (from project root)
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down

# Check migration version
go run cmd/migrate/main.go version
```

**Environment Variables for Database:**

- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_HOST` - Database host
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name

## Structure

Se tiene la siguiente estructura base, donde cada microservicio que forma parte del monolith se separa del resto para eventualmente escalar la arquitectura,
en cada uno los controladores, servicios y repositorios se ponen en sus correspontientes carpetas.

- **Crear su archivo separado para la interfaz y otro para la implementacion**.
- **No utilizar dependencias de un servicio en otro (no cross-imports)**.

```bash
.
├── cmd
│   └── migrate
├── config
├── docs
├── internal
│   ├── authentication-service
│   │   ├── acceptance-tests
│   │   │   └── features
│   │   ├── controllers
│   │   ├── repositories
│   │   ├── router
│   │   ├── services
│   │   └── utils
│   │       └── logger
│   ├── conversion-service
│   │   ├── controllers
│   │   ├── repositories
│   │   ├── router
│   │   ├── services
│   │   └── utils
│   │       └── logger
│   ├── router
│   ├── server
│   ├── utils
│   │   └── logger
│   └── world-browser-service
│       ├── controllers
│       ├── repositories
│       ├── router
│       ├── services
│       └── utils
│           └── logger
└── migrations
```

## Documentation

Endpoint documentation was made with `Swagger` <br>
Once staring up the project, go to this link to test out the endpoints:

```sh
# In this case, if locally starting the project, this is the url
http://localhost:8000/swagger/index.html
```
