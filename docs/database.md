# Database

Documentation for the database layer in `internal/database`.

---

## Overview

Formify Server uses PostgreSQL with the following stack:

- **pgx/v5** — PostgreSQL driver and connection pool
- **SQLC** — Type-safe SQL code generation

---

## Schema

### Users

| Column           | Type                       | Constraints               |
| ---------------- | -------------------------- | ------------------------- |
| `id`             | `SERIAL`                   | PRIMARY KEY               |
| `name`           | `VARCHAR(50)`              | NOT NULL                  |
| `email`          | `VARCHAR(100)`             | UNIQUE, NOT NULL          |
| `password`       | `VARCHAR(255)`             | NOT NULL                  |
| `oauth_provider` | `VARCHAR(50)`              |                           |
| `oauth_id`       | `VARCHAR(100)`             |                           |
| `is_oauth`       | `BOOLEAN`                  | DEFAULT FALSE             |
| `created_at`     | `TIMESTAMP WITH TIME ZONE` | DEFAULT CURRENT_TIMESTAMP |
| `updated_at`     | `TIMESTAMP WITH TIME ZONE` | DEFAULT CURRENT_TIMESTAMP |

### Forms

| Column        | Type                       | Constraints               |
| ------------- | -------------------------- | ------------------------- |
| `id`          | `SERIAL`                   | PRIMARY KEY               |
| `name`        | `VARCHAR(100)`             | NOT NULL                  |
| `description` | `TEXT`                     |                           |
| `user_id`     | `INTEGER`                  | NOT NULL                  |
| `status`      | `form_status`              | DEFAULT 'draft'           |
| `schema`      | `JSONB`                    | NOT NULL, DEFAULT '[]'    |
| `settings`    | `JSONB`                    | NOT NULL, DEFAULT '{}'    |
| `share_url`   | `TEXT`                     | UNIQUE                    |
| `created_at`  | `TIMESTAMP WITH TIME ZONE` | DEFAULT CURRENT_TIMESTAMP |
| `updated_at`  | `TIMESTAMP WITH TIME ZONE` | DEFAULT CURRENT_TIMESTAMP |

> **Note:** `form_status` is a custom ENUM type: `('draft', 'published')`

### Responses

| Column       | Type                       | Constraints               |
| ------------ | -------------------------- | ------------------------- |
| `id`         | `SERIAL`                   | PRIMARY KEY               |
| `form_id`    | `INTEGER`                  | NOT NULL                  |
| `data`       | `JSONB`                    | NOT NULL                  |
| `meta`       | `JSONB`                    | NOT NULL                  |
| `created_at` | `TIMESTAMP WITH TIME ZONE` | DEFAULT CURRENT_TIMESTAMP |

---

## Queries

### Users

| Query                | Type    | Description                       |
| -------------------- | ------- | --------------------------------- |
| `CreateUser`         | `:one`  | Create a new user                 |
| `CreateOAuthUser`    | `:one`  | Create a new OAuth user           |
| `GetUserByID`        | `:one`  | Get user by ID                    |
| `GetUserByEmail`     | `:one`  | Get user by email                 |
| `GetUserByOAuthID`   | `:one`  | Get user by OAuth provider and ID |
| `ListUsers`          | `:many` | List all users                    |
| `UpdateUser`         | `:one`  | Update user name/email            |
| `UpdateUserPassword` | `:exec` | Update user password              |
| `DeleteUser`         | `:exec` | Delete user by ID                 |

### Forms

| Query                        | Type    | Description                     |
| ---------------------------- | ------- | ------------------------------- |
| `CreateForm`                 | `:one`  | Create a new form               |
| `GetFormByID`                | `:one`  | Get form by ID                  |
| `GetFormByShareURL`          | `:one`  | Get form by share URL           |
| `ListFormsByUserID`          | `:many` | List all forms for a user       |
| `ListPublishedFormsByUserID` | `:many` | List published forms for a user |
| `UpdateForm`                 | `:one`  | Update form details             |
| `UpdateFormStatus`           | `:one`  | Update form status              |
| `UpdateFormShareURL`         | `:one`  | Update form share URL           |
| `DeleteForm`                 | `:exec` | Delete form by ID               |
| `CountFormsByUserID`         | `:one`  | Count forms for a user          |

### Responses

| Query                            | Type    | Description                     |
| -------------------------------- | ------- | ------------------------------- |
| `CreateResponse`                 | `:one`  | Create a new response           |
| `GetResponseByID`                | `:one`  | Get response by ID              |
| `ListResponsesByFormID`          | `:many` | List all responses for a form   |
| `ListResponsesByFormIDPaginated` | `:many` | List responses with pagination  |
| `DeleteResponse`                 | `:exec` | Delete response by ID           |
| `DeleteResponsesByFormID`        | `:exec` | Delete all responses for a form |
| `CountResponsesByFormID`         | `:one`  | Count responses for a form      |

---

## Connection

### `connection.go`

Manages the database connection pool.

| Function    | Description                                      |
| ----------- | ------------------------------------------------ |
| `InitDB()`  | Initializes connection pool using `DATABASE_URL` |
| `CloseDB()` | Gracefully closes the database connection pool   |

**Variables:** `DBPool *pgxpool.Pool` — Global connection pool instance

---

## Configuration

Set the connection string via environment variable:

```bash
DATABASE_URL=postgres://user:password@localhost:5432/formify?sslmode=disable
```

---

## SQLC

SQL queries are defined in `internal/database/queries/` and schema in `internal/database/schema/`.

Generated Go code is output to `internal/db/`.

```bash
make sqlc
```

See `internal/database/sqlc.yml` for SQLC settings.
