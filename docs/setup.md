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
| `PORT`         | Server port                  | `1323`  |

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
