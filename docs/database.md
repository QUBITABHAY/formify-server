# Database

Documentation for the database layer in `internal/database`.

---

## Overview

Formify Server uses PostgreSQL with the following stack:

- **pgx/v5** — PostgreSQL driver and connection pool
- **SQLC** — Type-safe SQL code generation

---

## Connection

### `connection.go`

Manages the database connection pool.

#### Functions

| Function    | Description                                      |
| ----------- | ------------------------------------------------ |
| `InitDB()`  | Initializes connection pool using `DATABASE_URL` |
| `CloseDB()` | Gracefully closes the database connection pool   |

#### Variables

- `DBPool *pgxpool.Pool` — Global connection pool instance

---

## Configuration

Set the connection string via environment variable:

```bash
DATABASE_URL=postgres://user:password@localhost:5432/formify?sslmode=disable
```

---

## SQLC

SQL queries are defined in `internal/database/queries/` and schema in `internal/database/schema/`.

### Generate Code

```bash
make sqlc
# or
sqlc generate
```

### Configuration

See `internal/database/sqlc.yml` for SQLC settings.

---

## Docker Compose

Start the database:

```bash
make db-up
```

Stop the database:

```bash
make db-down
```
