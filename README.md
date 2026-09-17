# Formify Server

Backend API for Formify, a form builder platform with support for:

- form creation and publishing
- response collection
- Google OAuth authentication
- optional Google Sheets integration
- file uploads

## Tech Stack

- Go (Echo v5)
- PostgreSQL
- sqlc for type-safe query generation
- golang-migrate for DB migrations
- Zap for structured logging

## Quick Start

1. Clone and install dependencies:

```bash
go mod download
```

2. Configure environment:

```bash
cp .env.example .env
```

3. Start local database:

```bash
make db-up
```

4. Run migrations:

```bash
make migrate-up
```

5. Start API:

```bash
make run
```

Server default URL: `http://localhost:1323`

## Development Commands

- `make run` - run API
- `make dev` - run with hot reload (requires `air`)
- `make build` - build binary to `bin/formify-server`
- `make test` - run tests
- `make lint` - run golangci-lint
- `make format` - format code
- `make sqlc` - regenerate DB query code
- `make migrate-up` / `make migrate-down` / `make migrate-status`

## Docker

To run the full stack (app + migrate + db):

```bash
docker compose up --build
```

This uses `docker-compose.yml` and exposes API on port `8080`.

## Documentation

- `docs/setup.md` - setup and local development
- `docs/api.md` - API endpoints and auth model
- `docs/database.md` - schema and migrations
- `docs/logging.md` - structured logging

## Changelog

See `CHANGELOG.md` for release notes and project history.
