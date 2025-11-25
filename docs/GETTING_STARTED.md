# Getting Started Guide

This guide will help you set up and run the OniChange POS System locally.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **PostgreSQL 15+** (optional if using Docker)
- **Redis 7+** (optional if using Docker)
- **Make** (optional, for convenience)

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd Go_Interview
```

### 2. Install Dependencies

```bash
make deps
# or
go mod download
```

### 3. Configure Environment

Copy the example environment file and update with your values:

```bash
cp .env.example .env
```

Edit `.env` and set at minimum:
- `JWT_ACCESS_SECRET` (minimum 32 characters)
- `JWT_REFRESH_SECRET` (minimum 32 characters)
- Database credentials if not using Docker

### 4. Start Infrastructure Services

Using Docker Compose:

```bash
cd deployments/docker
docker-compose up -d postgres redis
```

Or manually:
- Start PostgreSQL on port 5432
- Start Redis on port 6379

### 5. Run Database Migrations

```bash
# Install golang-migrate if not installed
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
make migrate-up
```

### 6. Build and Run Services

```bash
# Build all services
make build

# Run API Gateway (in one terminal)
./bin/api-gateway

# Run User Service (in another terminal)
./bin/user-service

# Run Order Service (in another terminal)
./bin/order-service
```

Or use Docker Compose for all services:

```bash
cd deployments/docker
docker-compose up
```

## Verify Installation

### Check Health Endpoints

```bash
# API Gateway
curl http://localhost:8080/health

# User Service
curl http://localhost:8082/health

# Order Service
curl http://localhost:8081/health
```

### Test API Gateway

```bash
# Health check
curl http://localhost:8080/health

# API endpoint (will return placeholder)
curl http://localhost:8080/api/v1/auth/login
```

## Development Workflow

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration
```

### Code Quality

```bash
# Lint code
make lint

# Security scan
make security-scan
```

### Hot Reload (Development)

For hot reload during development, use tools like:
- [Air](https://github.com/cosmtrek/air)
- [CompileDaemon](https://github.com/githubnemo/CompileDaemon)

Example with Air:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air -c .air.toml
```

## Project Structure

```
.
├── cmd/              # Service entry points
├── pkg/              # Shared packages
├── internal/         # Service-specific code
├── migrations/       # Database migrations
├── deployments/      # Deployment configs
└── tests/            # Test files
```

## Common Issues

### Port Already in Use

If you get "port already in use" errors:
- Change ports in `.env` file
- Or stop existing services using those ports

### Database Connection Failed

- Ensure PostgreSQL is running
- Check database credentials in `.env`
- Verify network connectivity

### Redis Connection Failed

- Ensure Redis is running
- Check Redis configuration in `.env`
- Verify Redis password if set

## Next Steps

- Read the [Architecture Documentation](architecture/ARCHITECTURE.md)
- Review [API Documentation](api/API.md)
- Check [Security Guidelines](security/SECURITY.md)

## Getting Help

- Check existing issues on GitHub
- Review documentation in `docs/`
- Contact the development team

