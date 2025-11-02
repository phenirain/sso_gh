# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an SSO (Single Sign-On) service built in Go that serves as an API Gateway for a multi-role e-commerce system. It handles authentication via JWT tokens and proxies HTTP requests to multiple backend gRPC services (Admin, Client, Manager).

**Tech Stack:**
- Go 1.25
- Echo web framework (HTTP REST API)
- PostgreSQL (via sqlx)
- gRPC clients for backend communication
- JWT authentication with access/refresh tokens
- Swagger documentation
- Docker containerization

## Key Architecture

### Service Role
The SSO service acts as an **API Gateway** that:
1. Handles user authentication (login, signup, token refresh)
2. Validates JWT tokens for all incoming requests
3. Proxies authenticated requests to appropriate backend gRPC services based on user role
4. Manages CORS and middleware layers

### Backend Communication Pattern
The service connects to **three separate gRPC backend services**:
- **Admin service** (`cfg.GRPC.Admin`): Client management, products, orders, reports
- **Client service** (`cfg.GRPC.Client`): Client profile, product browsing, ordering
- **Manager service** (`cfg.GRPC.Manager`): Order fulfillment and management

gRPC clients are initialized in `internal/application/server.go:50-84` and used throughout the application handlers.

### Authentication Flow
1. User logs in via `/auth/logIn` or signs up via `/auth/signUp`
2. Auth service (`internal/services/auth`) validates credentials and generates JWT tokens
3. Returns both access token (60min TTL, see `internal/run.go:48`) and refresh token
4. Subsequent requests include `Authorization: Bearer <token>` header
5. JWT middleware (`pkg/echomiddleware/jwt.go`) validates tokens and extracts `userId` into request context
6. Handlers use `contextkeys.UserIDCtxKey` to retrieve authenticated user ID

### Layer Structure
```
cmd/sso/main.go              # Entry point
internal/run.go              # Server initialization and lifecycle
internal/application/        # HTTP handlers (Echo routes)
  ├── auth/                  # Authentication endpoints
  ├── admin/                 # Admin role endpoints (proxy to gRPC)
  ├── client/                # Client role endpoints (proxy to gRPC)
  └── manager/               # Manager role endpoints (proxy to gRPC)
internal/services/           # Business logic layer
  └── auth/                  # Auth service implementation
internal/repository/         # Database access layer
  └── user/                  # User repository (PostgreSQL)
internal/domain/             # Domain models (User entity)
internal/lib/jwt/            # JWT token generation/parsing
pkg/                         # Shared utilities
  ├── echomiddleware/        # Echo middleware (JWT validation, logging, etc.)
  ├── database/              # DB connection helper
  └── logger/                # Structured logging setup
```

## Configuration

Configuration is loaded from `config/config.yaml` using Viper:
- `env`: Environment (dev/prod)
- `connection_string`: PostgreSQL connection
- `secret`: JWT signing secret
- `allowed_origins`: CORS whitelist
- `http.port`: HTTP server port (default 8081)
- `grpc.admin/client/manager`: Backend gRPC service addresses

See `internal/config/config.go` for full schema.

## Common Commands

### Building and Running
```bash
# Build binary
go build -o sso ./cmd/sso

# Run locally (requires config/config.yaml)
./sso

# Run with Go
go run ./cmd/sso
```

### Swagger Documentation
```bash
# Generate Swagger docs from annotations
make swagger_gen

# Access Swagger UI at http://localhost:8081/swagger/index.html
```

### Docker
```bash
# Build Docker image
make docker_build

# Push to registry
make docker_push

# Build and push
make docker_deploy

# Run with docker-compose
make docker_compose_up
make docker_compose_down
```

### Testing
Currently no test files exist in the project. To add tests, create `*_test.go` files and run:
```bash
go test ./...
go test -v ./internal/services/auth  # Test specific package
```

### Database
The application expects a PostgreSQL database. Connection is initialized in `pkg/database` via `MustInitDb()` which panics on failure. Ensure:
- Connection string in `config.yaml` is correct
- Database schema is created (not managed by this service)
- User repository expects tables: `users`, `audit` (with `user_id` foreign key)

## Development Notes

### Adding New Endpoints
1. Create handler in `internal/application/<role>/` directory
2. Implement gRPC client call if needed
3. Register route in `internal/application/server.go` (e.g., `registerAdminRoutes`)
4. Add Swagger annotations in handler function
5. Run `make swagger_gen` to update docs
6. Update JWT middleware skip list in `pkg/echomiddleware/jwt.go` if endpoint should be public

### JWT Middleware Bypass
Public endpoints (no auth required) are defined in `pkg/echomiddleware/jwt.go:17-23`:
- `/auth/*` routes
- `/health`
- `/swagger/*`

To add more public routes, update the `skip` map.

### Proto Files
gRPC protocol buffers are maintained in external repo:
`gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto`

Update `go.mod` version to pull new proto definitions.

### Error Handling
- Custom errors defined in `internal/errors/auth` and `internal/errors/jwt`
- Echo returns standard HTTP errors (`echo.ErrUnauthorized`, etc.)
- Use `internal/dto/response.ApiResponse[T]` for consistent JSON responses

### Server Lifecycle
- Uses `errgroup` for graceful shutdown (`internal/run.go:32-44`)
- HTTP server runs on port from config (`cfg.HTTP.Port`)
- Pprof server always runs on port 6060 for profiling
- SIGINT triggers graceful shutdown with 5s timeout
