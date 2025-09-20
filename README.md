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
# Run all tests in subdirectories and show coverage
go test ./... -cover
```

## Structure

Se tiene la siguiente estructura base, donde cada microservicio que forma parte del monolith se separa del resto para eventualmente escalar la arquitectura,
en cada uno los controladores, servicios y repositorios se ponen en sus correspontientes carpetas.

- **Crear su archivo separado para la interfaz y otro para la implementacion**.
- **No utilizar dependencias de un servicio en otro (no cross-imports)**.

```bash
.
├── cmd
│   └── main.go # Binary entrypoint
├── config
│   └── config.go # Server/Global config
├── Dockerfile
├── go.mod
├── go.sum
├── internal # Server logic separated by service and packaged by layer
│   ├── authentication-service # Auth Microservice
│   │   ├── controllers
│   │   ├── repositories
│   │   ├── router
│   │   ├── services
│   │   └── utils
│   ├── conversion-service # Conversions Microservice
│   │   ├── controllers
│   │   ├── repositories
│   │   ├── router
│   │   ├── services
│   │   └── utils
│   ├── router
│   │   └── router.go
│   ├── server
│   │   └── server.go
│   ├── utils
│   │   └── logger
│   └── world-browser-service # World Browser Microservice
│       ├── controllers
│       ├── repositories
│       ├── router
│       ├── services
│       └── utils

```
