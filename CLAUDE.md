# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go web application skeleton built with a clean architecture pattern using Gin framework. It's designed as a template for building scalable Go applications with best practices and modern tooling.

### Key Technologies
- **Web Framework**: Gin (v1.10.1)
- **Database**: GORM with MySQL/PostgreSQL support
- **Cache**: Redis (v9.11.0)
- **Message Queue**: RabbitMQ (amqp091-go v1.10.0)
- **Dependency Injection**: Google Wire (v0.6.0)
- **Configuration**: Viper (v1.20.1)
- **Logging**: Zap (v1.27.0)
- **Validation**: Validator (v10.27.0)
- **JWT**: golang-jwt/jwt (v5.2.2)

## Common Development Commands

### Build and Run
```bash
# Build all services (API, Consumer, Scheduler)
make build

# Build specific service
make api        # Build API service
make consumer   # Build consumer service
make scheduler  # Build scheduler service

# Run services
make run            # Run API service
make run-consumer   # Run consumer service
make run-scheduler  # Run scheduler service
```

### Development Tools
```bash
# Generate dependency injection code
make wire

# Code quality
make fmt      # Format code
make lint     # Run golangci-lint
make vet      # Run go vet
make test     # Run tests
make test-coverage  # Run tests with coverage

# Dependency management
make deps    # Tidy and download dependencies
make tools   # Install development tools
```

### Database Operations
```bash
# Database migration and seeding
make migrate     # Run database migrations
make seed        # Run database seeding
make db-reset    # Reset database (migrate + seed)
```

### Docker Development
```bash
# Start development environment
make up          # Start Docker services
make down        # Stop Docker services
make logs        # View service logs
make shell       # Enter API container
make db-shell    # Enter database container
```

## Architecture Overview

### Dependency Injection Pattern
The project uses Google Wire for compile-time dependency injection. Key concepts:
- **Provider Sets**: Organized by layer (Infrastructure, Repository, Service, Handler)
- **Direct Dependencies**: App struct contains all dependencies directly (no wrapper pattern)
- **Generated Code**: `internal/wire/wire_gen.go` is auto-generated, don't edit manually

### Application Structure
```
cmd/           # Application entry points
├── api/       # HTTP API server
├── consumer/  # Message queue consumer
└── scheduler/ # Job scheduler

internal/      # Private application code
├── app/       # Application container and lifecycle
├── config/    # Configuration management
├── handler/   # HTTP handlers (v1/)
├── service/   # Business logic layer
├── repository/ # Data access layer
├── model/     # Data models
├── router/    # Route definitions
├── middleware/ # HTTP middleware
├── messaging/ # Message queue handling
├── scheduler/ # Job scheduling
└── wire/      # Dependency injection configuration

pkg/           # Public packages
├── database/  # Database connection utilities
├── redis/     # Redis client utilities
├── mq/        # RabbitMQ utilities
├── logger/    # Logging utilities
├── jwt/       # JWT utilities
├── bcrypt/    # Password hashing
├── validator/ # Input validation
└── idgen/     # ID generation (Sonyflake)
```

### Dependency Flow
```
Handler → Service → Repository → Database
    ↓           ↓           ↓
  Router    Messaging   Infrastructure
```

## Configuration

### Environment-based Configuration
- Development: `configs/config.dev.yaml`
- Docker: `configs/config.docker.yaml`
- Production: `configs/config.prod.yaml`

### Key Configuration Sections
- **App**: Server settings (host, port, environment)
- **Databases**: Multi-database support (MySQL/PostgreSQL)
- **Redis**: Cache configuration
- **RabbitMQ**: Message queue with exchanges/queues
- **Scheduler**: Job scheduling configuration
- **JWT**: Authentication settings
- **Logger**: Zap logger configuration

### Environment Variables
Configuration supports environment variable overrides:
```bash
APP_PORT=8081          # Override server port
DATABASES_DEFAULT_DSN=...  # Override database DSN
REDIS_ADDR=localhost:6379   # Override Redis address
```

## Database Architecture

### Multi-Database Support
- Primary database configured as "default"
- Additional databases can be configured in `databases` section
- GORM models support automatic migrations

### Models and Repositories
- Models defined in `internal/model/`
- Repositories in `internal/repository/`
- Services use repositories for data access

### Migration Strategy
- Migration scripts in `scripts/migrate/`
- Seeding scripts in `scripts/seed/`
- Use `make migrate` and `make seed` commands

## Message Queue Architecture

### RabbitMQ Integration
- Producer-consumer pattern
- Configurable exchanges and queues
- Message processors in `internal/messaging/processors/`

### Message Flow
1. HTTP handlers publish messages to RabbitMQ
2. Consumer service processes messages asynchronously
3. Each message type has dedicated processor

## Adding New Features

### New API Endpoint
1. Add model in `internal/model/`
2. Add repository in `internal/repository/`
3. Add service in `internal/service/`
4. Add handler in `internal/handler/v1/`
5. Add route in `internal/router/api/v1/`
6. Update Wire providers in `internal/wire/providers.go`
7. Run `make wire` to regenerate DI code

### New Message Handler
1. Add processor in `internal/messaging/processors/`
2. Configure queue in config file
3. Register processor in consumer

### New Scheduled Job
1. Add job in `internal/scheduler/jobs/`
2. Configure job in `scheduler` section of config
3. Job automatically registered with scheduler

## Testing

### Test Commands
```bash
make test            # Run all tests
make test-coverage   # Run tests with coverage report
```

### Testing Patterns
- Unit tests for individual components
- Integration tests using Wire-generated dependencies
- Mock external dependencies for isolated testing

## Docker Development

### Development Environment
- All services configured in `docker-compose.yaml`
- Override settings in `docker-compose.override.yaml`
- Production settings in `docker-compose.prod.yaml`

### Service Ports
- API: 8080
- PostgreSQL: 5432
- Redis: 6379
- RabbitMQ: 5672 (AMQP), 15672 (Management UI)
- Adminer: 8081 (Database management)

## Important Notes

### Wire Dependency Injection
- Always run `make wire` after modifying dependencies
- Don't edit `internal/wire/wire_gen.go` directly
- Provider sets are organized by architectural layer

### Configuration Management
- Use environment variables for sensitive data
- Configuration files for structured settings
- Viper handles merging file config with environment variables

### Code Style
- Follow existing Go conventions in the codebase
- Use golangci-lint for code quality checks
- Format code with `make fmt` before committing

### Multi-Service Architecture
- Three main services: API, Consumer, Scheduler
- Each can be built and run independently
- Shared configuration and dependencies through Wire