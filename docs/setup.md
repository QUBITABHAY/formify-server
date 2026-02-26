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

| Variable       | Description                  | Default |
| -------------- | ---------------------------- | ------- |
| `DATABASE_URL` | PostgreSQL connection string | —       |
| `JWT_SECRET`   | JWT signing secret           | —       |
| `PORT`         | Server port                  | `1323`  |

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
make db-migrate
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
| `make db-up`   | Start database                     |
| `make db-down` | Stop database                      |
| `make sqlc`    | Generate code from SQL             |
