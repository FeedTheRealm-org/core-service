# Core Service Architecture

## Overview

The Feed the Realm core-service is currently designed as a **modular monolith** with the future goal of easily splitting it into a microservices architecture. It uses the [Gin-gonic](https://gin-gonic.com/) framework for routing and HTTP handling.

## Directory Structure

```
├── cmd
│   └── main.go                 # Entry point of the application
├── config
│   ├── config.go               # Environment configurations
│   └── database.go             # Database connection setup
├── internal                    # Private application code
│   ├── assets-service          # Domain logic for game assets
│   ├── authentication-service  # Domain logic for auth and users
│   ├── players-service         # Domain logic for players and characters
│   ├── world-service           # Domain logic for game worlds and Nomad orchestration
│   ├── common_handlers         # Shared HTTP handlers
│   ├── errors                  # Shared custom error types
│   ├── middleware              # Gin middlewares (e.g., JWT, Error handler)
│   ├── router                  # Global router setup
│   ├── server                  # Server lifecycle management
│   └── utils                   # Shared utilities (logging, validation)
└── migrations                  # Database migration scripts
```

## Services Division

The `internal` directory isolates each domain into its own "service" package (`assets-service`, `authentication-service`, `players-service`, `world-service`).

Each service independently contains its own:

- **controllers**: HTTP handlers
- **services**: Business logic
- **repositories**: Database interactions
- **models**: Domain entities
- **dtos**: Data Transfer Objects for requests/responses
- **router**: Service-specific route registration
- **acceptance-tests**: Domain-specific tests

### Rules for Future Proofing

To guarantee an easy transition to a microservices architecture later on:

1. **Interface Driven:** Create separate files for the interface and the implementation.
2. **No Cross-Imports:** Do not import dependencies directly from one service to another. If they need to communicate, it should be done over APIs, events, or shared common libraries if absolutely necessary.
