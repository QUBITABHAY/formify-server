# Setup Guide

Getting started with Formify Server.

---

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

---

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/formify-server.git
cd formify-server
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Install Dev Tools (Optional)

```bash
make install-dev-tools
```

This installs:

- **air** — Hot reload for development
- **golangci-lint** — Linter
- **sqlc** — SQL code generator

---

## Environment Variables

Copy the example environment file:

```bash
cp .env.example .env
```

### Required Variables

| Variable       | Description                                                                   | Default       |
| -------------- | ----------------------------------------------------------------------------- | ------------- |
| `DATABASE_URL` | PostgreSQL connection string                                                  | -             |
| `JWT_SECRET`   | JWT signing secret                                                            | -             |
| `PORT`         | Server port                                                                   | `1323`        |
| `ENV`          | Runtime mode (`development` = console logs, `production` = JSON logs)        | `development` |

### Google/Auth Variables

| Variable | Description | Default |
| --- | --- | --- |
| `SESSION_SECRET` | Session secret used by OAuth flow | `formify-session-secret` |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | — |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | — |
| `GOOGLE_CALLBACK_URL` | OAuth callback URL | `http://localhost:1323/api/auth/google/callback` |
| `FRONTEND_URL` | Frontend base URL used for OAuth callback redirect | `http://localhost:5173` |
| `CORS_ORIGINS` | Comma-separated allowed origins | falls back to `FRONTEND_URL` |

### Optional Google Sheets Service Account

| Variable | Description |
| --- | --- |
| `GOOGLE_SERVICE_ACCOUNT_KEY_PATH` | Path to service account JSON key used as fallback for Sheets operations |
| `GOOGLE_SERVICE_ACCOUNT_KEY` | Raw JSON service account key (alternative to file path) |

### Optional Upload Variables

| Variable | Description |
| --- | --- |
| `CLOUDINARY_CLOUD_NAME` | Cloudinary cloud name for file upload storage |
| `CLOUDINARY_API_KEY` | Cloudinary API key |
| `CLOUDINARY_API_SECRET` | Cloudinary API secret |

Notes:

- If a user logs in with Google, their OAuth token is used for Sheets operations.
- If user token is unavailable or expired, users must log in with Google again to use Sheets operations.

---

## Running the Server

### Development

Start with hot reload:

```bash
make dev
```

### Standard

```bash
make run
```

### Build Binary

```bash
make build
./bin/formify-server
```

---

## Database Setup

### Start PostgreSQL with Docker

```bash
make db-up
```

### Run Migrations

```bash
make migrate-up
```

Other migration helpers:

```bash
make migrate-down
make migrate-reset
make migrate-status
```

### Stop Database

```bash
make db-down
```

---

## Available Commands

Run `make help` to see all available commands:

| Command        | Description                        |
| -------------- | ---------------------------------- |
| `make run`     | Run the server                     |
| `make build`   | Build the server binary            |
| `make dev`     | Run with hot reload (requires air) |
| `make test`    | Run all tests                      |
| `make lint`    | Run linter                         |
| `make format`  | Format code                        |
| `make vet`     | Run `go vet`                       |
| `make db-up`   | Start database                     |
| `make db-down` | Stop database                      |
| `make migrate-up` | Apply all pending migrations    |
| `make migrate-down` | Roll back last migration     |
| `make migrate-status` | Show current migration version |
| `make sqlc`    | Generate code from SQL             |

## Docker Compose Option

Run the app, migrations, and PostgreSQL together:

```bash
docker compose up --build
```

In this mode, API is exposed on `http://localhost:8080`.
