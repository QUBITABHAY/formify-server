# CLAUDE.md — Formify Server

This document provides architecture context, development guidelines, schema definitions, and command references for Claude Code when working in the `formify-server` Go backend repository.

---

## 1. Project Overview

**Formify Server** is a high-performance REST API backend written in Go (Echo v5) that powers the Formify form builder platform. It provides endpoints for user authentication, form creation and publishing, response collection and querying, Google OAuth, Google Sheets automated synchronization, and Cloudinary media uploads.

---

## 2. Tech Stack

- **Language**: [Go](https://go.dev/) (1.25+)
- **HTTP Framework**: [Echo v5](https://echo.labstack.com/) (`github.com/labstack/echo/v5`)
- **Database**: [PostgreSQL](https://www.postgresql.org/) with `pgx/v5` connection pool (`github.com/jackc/pgx/v5`, `github.com/jackc/puddle/v2`)
- **SQL Generator**: [sqlc](https://sqlc.dev/) for compile-time type-safe Go query generation (`internal/database/sqlc.yml`)
- **Database Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate) (`github.com/golang-migrate/migrate/v4`)
- **Structured Logging**: [Uber Zap](https://github.com/uber-go/zap) (`go.uber.org/zap`)
- **Authentication**: JWT (`github.com/golang-jwt/jwt/v5`), Google OAuth via Goth (`github.com/markbates/goth`, `golang.org/x/oauth2`)
- **Third-Party Integrations**:
  - Google Sheets API v4 (`google.golang.org/api/sheets/v4`)
  - Cloudinary Go SDK (`github.com/cloudinary/cloudinary-go/v2`)
- **Configuration**: [Viper](https://github.com/spf13/viper) with `.env` file parsing

---

## 3. Development Commands

Run these commands from the `formify-server/` root directory:

### Server Lifecycle
```bash
# Run server directly (starts on port configured in .env or default 1323)
make run
# or: go run ./cmd/api

# Run server with live reload (requires air)
make dev

# Compile production binary into bin/formify-server
make build

# Clean up binaries and temporary build files
make clean
```

### Code Quality & Testing
```bash
# Run all Go tests
make test

# Run tests with verbose output
make test-v

# Run linter across all packages (requires golangci-lint)
make lint

# Format code with standard gofmt
make format

# Run Go static analysis
make vet
```

### Database & Migrations
```bash
# Start local PostgreSQL database container using Docker Compose
make db-up

# Stop PostgreSQL container
make db-down

# Apply all pending database migrations
make migrate-up
# or: go run ./cmd/migrate -up

# Roll back the most recent migration step
make migrate-down
# or: go run ./cmd/migrate -down

# Check current database migration version and dirty status
make migrate-status
# or: go run ./cmd/migrate -version

# Reset database (rollback all migrations)
make migrate-reset
# or: go run ./cmd/migrate -reset

# Create a new versioned migration up/down pair
make migrate-create

# Generate type-safe Go database code from SQL via sqlc
make sqlc
# or: cd internal/database && sqlc generate
```

### Development Tools Installation
```bash
# Install air, golangci-lint, and sqlc CLI tools
make install-dev-tools
```

---

## 4. Architecture & Directory Map

```text
formify-server/
├── cmd/
│   ├── api/                           # Application entrypoint
│   │   └── main.go                    # Dependency injection, middleware setup, route registration
│   └── migrate/                       # CLI migration tool
│       └── main.go                    # Golang-migrate runner supporting -up, -down, -reset, -version, -steps, -force
├── docs/                              # Project documentation
│   ├── api.md                         # API endpoints and request/response specifications
│   ├── database.md                    # Database schema, table structures, and queries
│   ├── logging.md                     # Structured logging conventions
│   └── setup.md                       # Setup and local environment guide
├── infrastructure/                    # Container orchestration
│   └── docker-compose.yml             # Local PostgreSQL database service
├── internal/                          # Private application packages
│   ├── auth/                          # Authentication domain
│   │   ├── google_login.go            # Google OAuth flow & token exchange
│   │   ├── handler.go                 # Google login/callback & logout HTTP handlers
│   │   ├── providers.go               # Goth OAuth provider configuration
│   │   └── service.go                 # JWT generation & authentication service
│   ├── config/                        # Configuration loader
│   │   └── config.go                  # Viper loader for environment variables
│   ├── database/                      # Database configuration and source files
│   │   ├── connection.go              # PostgreSQL connection pool initializer
│   │   ├── migrations/                # Versioned SQL migration files:
│   │   │   ├── 000001_init_schema.up.sql / .down.sql
│   │   │   ├── 000002_add_google_sheets.up.sql / .down.sql
│   │   │   └── 000003_add_oauth_tokens.up.sql / .down.sql
│   │   ├── queries/                   # SQL queries used by sqlc:
│   │   │   ├── forms.sql              # Form queries (CRUD, share URL, publish)
│   │   │   ├── responses.sql          # Response queries (CRUD, pagination, counts)
│   │   │   └── users.sql              # User & OAuth token queries
│   │   ├── schema/                    # Source SQL schema files for sqlc:
│   │   │   ├── forms.sql              # forms table definition & form_status enum
│   │   │   ├── responses.sql          # responses table definition
│   │   │   └── users.sql              # users table definition
│   │   └── sqlc.yml                   # SQLC configuration file
│   ├── db/                            # Auto-generated Go code by sqlc (DO NOT EDIT MANUALLY)
│   │   ├── db.go                      # DBTX interface and DB connection wrappers
│   │   ├── forms.sql.go               # Generated form query methods
│   │   ├── models.go                  # Generated Go structs for tables and enums
│   │   ├── querier.go                 # Querier interface
│   │   ├── responses.sql.go           # Generated response query methods
│   │   └── users.sql.go               # Generated user query methods
│   ├── file_upload/                   # File uploads domain (Cloudinary)
│   │   ├── file.go                    # Cloudinary client and uploader service
│   │   └── handler.go                 # Multipart file upload handler
│   ├── form/                          # Form domain
│   │   ├── handler.go                 # Form CRUD, publish/unpublish, and Sheets handlers
│   │   ├── model.go                   # Form domain models and DTOs
│   │   ├── repository.go              # Form database repository
│   │   └── service.go                 # Form business logic & FormGetterAdapter
│   ├── integrations/
│   │   └── google/                    # Google external integrations
│   │       ├── service.go             # Google OAuth token refresher & credential management
│   │       └── sheets.go              # Google Sheets API client & spreadsheet sync
│   ├── logger/                        # Structured logging
│   │   ├── logger.go                  # Zap logger initialization and helper functions
│   │   └── middleware.go              # Echo HTTP request logging middleware
│   ├── middleware/                    # HTTP middleware
│   │   └── auth.go                    # JWT authentication middleware (Cookie + Bearer header)
│   ├── response/                      # Response domain
│   │   ├── handler.go                 # Response submission and query handlers
│   │   ├── model.go                   # Response domain models and DTOs
│   │   ├── repository.go              # Response database repository
│   │   └── service.go                 # Response service with automatic Sheets sync
│   ├── shared/                        # Shared utilities
│   │   ├── handler.go                 # Healthcheck endpoints (/health, /health/db) & RespondError
│   │   └── helpers.go                 # pgtype conversion helpers, GetAuthUserID, GenerateShareURL
│   └── user/                          # User domain
│       ├── handler.go                 # User creation and lookup handlers
│       ├── model.go                   # User domain models and DTOs
│       ├── repository.go              # User database repository
│       └── service.go                 # User business logic service
├── .air.toml                          # Live reload configuration for air
├── .env.example                       # Example environment variables
├── .golangci.yml                      # GolangCI-Lint configuration
├── Dockerfile                         # Production container build
├── go.mod                             # Go module dependencies
├── go.sum                             # Go module checksums
└── Makefile                           # Build, test, and migration automation targets
```

---

## 5. API Endpoints Map

### Public Endpoints
- `GET /` — Basic status probe
- `GET /health` — Service health check
- `GET /health/db` — Database connectivity health check
- `POST /api/users` — Register a new user
- `GET /api/forms/share/:share_url` — Get published form schema by share URL
- `POST /api/forms/:form_id/responses` — Submit responses to a published form
- `POST /api/forms/:form_id/upload` — Upload file attachment (Cloudinary)
- `POST /api/auth/logout` — Clear auth cookie
- `GET /api/auth/google` — Initiate Google OAuth login
- `GET /api/auth/google/callback` — Google OAuth redirect receiver

### Protected Endpoints (Requires valid JWT via cookie or `Authorization: Bearer <token>`)
- `GET /api/auth/me` — Get current authenticated user
- `GET /api/users/:id` — Get user profile (self only)
- `GET /api/users/:id/forms` — Get all forms for user
- `POST /api/forms` — Create a new form
- `GET /api/forms/:id` — Get form details (owner only)
- `PUT /api/forms/:id` — Update form details, schema, or settings (owner only)
- `DELETE /api/forms/:id` — Delete a form (owner only)
- `POST /api/forms/:id/publish` — Publish form and generate share URL (owner only)
- `POST /api/forms/:id/unpublish` — Unpublish form (owner only)
- `POST /api/forms/:id/sheets/create` — Create and link a Google Spreadsheet (owner only)
- `DELETE /api/forms/:id/sheets/link` — Unlink Google Spreadsheet (owner only)
- `GET /api/forms/:id/responses` — List responses for a form (owner only)
- `GET /api/responses/:id` — Get single response details (owner only)
- `DELETE /api/responses/:id` — Delete a response (owner only)

---

## 6. Database Schema & Migration Workflow

### Canonical Schema Summary
1. **`users` table** (`internal/database/schema/users.sql`): `id`, `name`, `email`, `password`, `oauth_provider`, `oauth_id`, `is_oauth`, `google_access_token`, `google_refresh_token`, `google_token_expiry`, timestamps.
2. **`forms` table** (`internal/database/schema/forms.sql`): `id`, `form_id`, `name`, `description`, `user_id`, `status` (`'draft' | 'published'`), `schema` (JSONB), `settings` (JSONB), `share_url`, `google_sheet_id`, `google_sheet_name`, `google_sheet_linked_at`, `google_sheet_auto_sync`, timestamps.
3. **`responses` table** (`internal/database/schema/responses.sql`): `id`, `form_id`, `data` (JSONB), `meta` (JSONB), `created_at`.

### Step-by-Step Schema Modification
1. Create migration files: `make migrate-create` (or create `<timestamp>_<name>.up.sql` and `*.down.sql` in `internal/database/migrations/`).
2. Write raw SQL DDL in `.up.sql` and `.down.sql`.
3. Update the matching table definition in `internal/database/schema/<domain>.sql`.
4. Update or add queries in `internal/database/queries/<domain>.sql`.
5. Run migrations: `make migrate-up`.
6. Regenerate Go code: `make sqlc` (runs `cd internal/database && sqlc generate`).

---

## 7. Environment Configuration Reference

Create `.env` based on `.env.example`:

```bash
# Server Configuration
PORT=1323
ENV=development

# Frontend & CORS Configuration
FRONTEND_URL=http://localhost:5173
CORS_ORIGINS=http://localhost:5173

# Database
DATABASE_URL=postgres://username:password@localhost:5432/formify?sslmode=disable

# Authentication & Session
JWT_SECRET=your-super-secret-jwt-key
SESSION_SECRET=your-session-secret

# Google OAuth & Sheets
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_CALLBACK_URL=http://localhost:1323/api/auth/google/callback
GOOGLE_SERVICE_ACCOUNT_KEY_PATH=/path/to/service-account-key.json

# Cloudinary Media Storage
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```
