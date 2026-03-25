# Production Usage

This document outlines the approach for building and deploying the core-service in a production environment.

## Docker Environments

The project uses a multi-stage `Dockerfile` to optimize the final artifact:

- `deps`: Downloads Go modules and installs tools (e.g., swaggo).
- `builder`: Compiles the Go application statically.
- `prod`: The final lightweight image containing only the compiled binary and necessary directories (`certs`, `migrations`, `templates`, etc.).

## Starting Production Server

The production environment uses the `prod` profile in docker-compose.

```bash
# Build the production image
make build

# Start the application
make up

# Alternatively, you can use the shorter command:
make up-build
```

## Environment Variables

Ensure all production secrets are correctly set in the `.env` file, particularly:

- `SERVER_ENVIRONMENT=production`
- The rest of the env vars are the same as development (if running locally with compose)
- Keep in mind buckets will still be local and nomad endpoints wont work!
- In real a deployment all the necessary env vars are passed via AWS SSM parameters and access to buckets, ecr, nomad, consul and others is handled via secure groups and IAM roles.

## Container Orchestration (Nomad)

The `world-service` is responsible for spinning up dynamic game servers using [HashiCorp Nomad](https://developer.hashicorp.com/nomad/docs).
The job template for the game servers is located at `internal/world-service/nomad/ftr-server-job.nomad`. Ensure a Nomad cluster is configured and accessible from the production core-service container.
