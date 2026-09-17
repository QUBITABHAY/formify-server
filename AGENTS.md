# AGENTS.md — Formify Server Developer & AI Agent Guide

This document is the authoritative guide for AI coding assistants and autonomous agents working on the **Formify Server** codebase (`formify-server/`).

---

## 1. Role & Mission

You are an expert Go Backend Engineer working on `formify-server`. Your responsibilities:
1. Maintain robust, high-performance, and idiomatic Go code adhering to Echo v5 patterns.
2. Strictly follow the 3-layer architecture (`Handler` ➔ `Service` ➔ `Repository` ➔ `sqlc`).
3. Manage database changes through versioned migrations (`golang-migrate`) and type-safe query generation (`sqlc`).
4. Ensure rigorous authentication, ownership validation, and error handling across all endpoints.

---

## 2. Technology Stack Reference

| Component | Library / Tool | Role |
| :--- | :--- | :--- |
| **Language** | Go 1.25+ | Server runtime and backend code |
| **HTTP Framework** | `github.com/labstack/echo/v5` | Router, parameter binding, middleware |
| **Database Driver** | `github.com/jackc/pgx/v5`, `pgxpool` | PostgreSQL connection pooling |
| **Query Generator** | `sqlc` | Type-safe Go query code from raw SQL |
| **Migrations** | `github.com/golang-migrate/migrate/v4` | Version-controlled database schema migrations |
| **Logging** | `go.uber.org/zap` | Structured JSON logging |
| **Auth** | `github.com/golang-jwt/jwt/v5`, `github.com/markbates/goth` | JWT authentication and Google OAuth 2.0 |
| **Third-Party APIs** | Google Sheets v4, Cloudinary SDK | Automated spreadsheet synchronization and media storage |

---

## 3. Project Structure & Dependency Flow

```text
cmd/
├── api/main.go               # Server initialization, dependency wiring, route mapping
└── migrate/main.go           # CLI tool for database migrations (-up, -down, -reset, -version, -steps, -force)

internal/
├── auth/                     # Google OAuth flow, JWT tokens, and login/logout handlers
│   ├── google_login.go       # Google OAuth flow & token exchange
│   ├── handler.go            # Google login/callback & logout HTTP handlers
│   ├── providers.go          # Goth OAuth provider configuration
│   └── service.go            # JWT generation & authentication service
├── config/                   # Viper configuration loading from .env and environment variables
│   └── config.go             # Config struct & Viper unmarshaling
├── database/                 # Database connection, migrations, queries, and schema:
│   ├── connection.go         # pgxpool connection management
│   ├── migrations/           # Versioned SQL migrations (*.up.sql / *.down.sql)
│   ├── queries/              # SQL queries for sqlc (forms.sql, responses.sql, users.sql)
│   ├── schema/               # Canonical SQL schemas (forms.sql, responses.sql, users.sql)
│   └── sqlc.yml              # SQLC compiler configuration
├── db/                       # Auto-generated Go code produced by sqlc (DO NOT EDIT DIRECTLY)
├── file_upload/              # Media upload handler and Cloudinary service
│   ├── file.go               # Cloudinary client and uploader service
│   └── handler.go            # Multipart file upload handler
├── form/                     # Form domain
│   ├── handler.go            # Form HTTP handlers
│   ├── model.go              # Form domain models and DTOs
│   ├── repository.go         # Form database repository
│   └── service.go            # Form business logic & FormGetterAdapter
├── integrations/google/      # Google Sheets client & OAuth token refresh logic
│   ├── service.go            # Google OAuth token refresher & credential management
│   └── sheets.go             # Google Sheets API client & spreadsheet sync
├── logger/                   # Zap logger initialization and Echo request logging middleware
│   ├── logger.go             # Zap logger initialization
│   └── middleware.go         # Echo request logger
├── middleware/               # Auth middleware (validates JWT from Cookie or Authorization header)
│   └── auth.go               # JWT validator middleware
├── response/                 # Response capture domain & automatic Google Sheets sync
│   ├── handler.go            # Response HTTP handlers
│   ├── model.go              # Response domain models and DTOs
│   ├── repository.go         # Response database repository
│   └── service.go            # Response business logic & Sheets sync
├── shared/                   # Shared utilities
│   ├── handler.go            # Healthcheck endpoints (/health, /health/db) & RespondError
│   └── helpers.go            # pgtype conversion helpers, GetAuthUserID, GenerateShareURL
└── user/                     # User domain
    ├── handler.go            # User HTTP handlers
    ├── model.go              # User domain models and DTOs
    ├── repository.go         # User database repository
    └── service.go            # User business logic service
```

---

## 4. Layered Architecture & Implementation Rules

Every feature domain in `internal/` must adhere strictly to the following layer responsibilities:

### 1. Handler Layer (`internal/<domain>/handler.go`)
- Receives `echo.Context`.
- Parses and binds query params, path params, and JSON request bodies into domain DTOs.
- Extracts authenticated user context: `c.Get("user_id").(float64)` or `c.Get("email").(string)`.
- Calls the corresponding Service methods.
- Formats and writes the HTTP response using standard JSON envelopes via `c.JSON()` or `shared.RespondError()`.
- **Rule**: Handlers must never directly query the database or invoke `*db.Queries`.

### 2. Service Layer (`internal/<domain>/service.go`)
- Implements core business rules, entity validations, and permission checks (e.g. confirming form ownership before update/delete).
- Coordinates cross-domain logic and third-party integrations (e.g. syncing new responses to Google Sheets).
- Interacts with the database exclusively through the Repository layer.
- **Rule**: Services must remain decoupled from HTTP constructs (`echo.Context`).

### 3. Repository Layer (`internal/<domain>/repository.go`)
- Interfaces directly with `*db.Queries` generated by sqlc.
- Converts sqlc database models to domain entities and vice versa.
- Handles database transactions and connection contexts.
- **Rule**: Repositories must never make external API calls or handle HTTP responses.

---

## 5. Step-by-Step Developer Workflows

### 5.1 Adding a New Endpoint
1. If new database queries are needed, add them to `internal/database/queries/<domain>.sql` and run `make sqlc`.
2. Add repository methods in `internal/<domain>/repository.go`.
3. Add business logic and service methods in `internal/<domain>/service.go`.
4. Add handler methods in `internal/<domain>/handler.go`.
5. Register the route in `cmd/api/main.go` under `api` (public) or `protected` (authenticated).

### 5.2 Modifying the Database Schema
1. Create a migration file pair using `make migrate-create` (or create `<timestamp>_<name>.up.sql` and `*.down.sql` in `internal/database/migrations/`).
2. Write clean PostgreSQL DDL in `.up.sql` and the inverse rollback in `.down.sql`.
3. Keep `internal/database/schema/<domain>.sql` in sync with the updated table definitions.
4. Add or update queries in `internal/database/queries/<domain>.sql`.
5. Apply the migration locally: `make migrate-up`.
6. Regenerate Go models: `make sqlc` (which runs `cd internal/database && sqlc generate`).

---

## 6. Error Handling & Structured Logging

- **Logging**: Always use structured logging with typed fields via `internal/logger`:
  ```go
  log := logger.GetLogger()
  log.Info("Form created", logger.ToField("form_id", form.ID), logger.ToField("user_id", userID))
  log.Error("Database query failed", logger.ToField("error", err))
  ```
- **Error Responses**: Return errors as JSON objects with an `"error"` message string:
  ```go
  return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
  ```

---

## 7. Verification Checklist for AI Agents

Before concluding any backend task, run and verify:
1. **Format Check**:
   ```bash
   make format
   ```
2. **Static Analysis**:
   ```bash
   make vet
   ```
3. **Linter Check**:
   ```bash
   make lint
   ```
4. **Test Suite**:
   ```bash
   make test
   ```
5. **Compilation Verification**:
   ```bash
   make build
   ```
   Ensures `bin/formify-server` builds cleanly without compilation errors.
