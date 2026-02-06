# API Documentation

Documentation for the Formify Server API endpoints.

---

## Base URL

The API runs on port `1323` by default. Configure via `PORT` environment variable.

```
http://localhost:1323
```

---

## Authentication

Most endpoints require JWT authentication. Include the JWT token in the `Authorization` header:

```
Authorization: Bearer <your-jwt-token>
```

Public endpoints (no authentication required):

- `POST /api/auth/login` - Email/password login
- `POST /api/users` - Create user account
- `POST /api/forms/:form_id/responses` - Submit form response
- `GET /api/auth/google` - Google OAuth login
- `GET /api/auth/google/callback` - Google OAuth callback

---

## Authentication API

### Login

**POST** `/api/auth/login`

Login with email and password. Returns JWT token for authenticated requests.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:** `200 OK`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "user@example.com"
  }
}
```

**Error:** `401 Unauthorized`

```json
{
  "error": "Invalid email or password"
}
```

---

### Google OAuth Login

**GET** `/api/auth/google`

Initiates Google OAuth flow. Redirects to Google login page.

---

### Google OAuth Callback

**GET** `/api/auth/google/callback`

Handles OAuth callback from Google. Creates/updates user and returns JWT token.

**Response:** `200 OK`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "user@example.com"
  }
}
```

---

## Users API

### Create User

**POST** `/api/users`

Creates a new user account.

**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secure_password"
}
```

**Response:** `201 Created`

```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com"
}
```

---

### Get User

**GET** `/api/users/:id` 🔒

Retrieves a user by ID. Requires authentication.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com"
}
```

---

## Forms API

### Create Form

**POST** `/api/forms` 🔒

Creates a new form with default draft status. Requires authentication.

**Request Body:**

```json
{
  "name": "Customer Survey",
  "description": "Optional description",
  "user_id": 1,
  "schema": [],
  "settings": {}
}
```

**Response:** `201 Created`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "description": "Optional description",
  "user_id": 1,
  "status": "draft",
  "schema": [],
  "settings": {},
  "share_url": null,
  "created_at": "2024-01-28T12:00:00Z",
  "updated_at": "2024-01-28T12:00:00Z"
}
```

---

### Get Form

**GET** `/api/forms/:id` 🔒

Retrieves a form by ID. Requires authentication.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "description": "Optional description",
  "user_id": 1,
  "status": "draft",
  "schema": [],
  "settings": {},
  "share_url": null,
  "created_at": "2024-01-28T12:00:00Z",
  "updated_at": "2024-01-28T12:00:00Z"
}
```

**Error:** `404 Not Found`

```json
{
  "error": "Form not found"
}
```

---

### Get User Forms

**GET** `/api/users/:id/forms` 🔒

Retrieves all forms belonging to a user. Requires authentication.

**Response:** `200 OK`

```json
[
  {
    "id": 1,
    "name": "Customer Survey",
    "user_id": 1,
    "status": "draft",
    "schema": [],
    "settings": {},
    ...
  }
]
```

---

### Update Form

**PUT** `/api/forms/:id` 🔒

Updates form details. Requires authentication.

**Request Body:**

```json
{
  "name": "Updated Survey Name",
  "description": "Updated description",
  "schema": [...],
  "settings": {...}
}
```

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Updated Survey Name",
  "description": "Updated description",
  ...
}
```

---

### Publish Form

**POST** `/api/forms/:id/publish` 🔒

Sets a form's status to `published`. Requires authentication.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "status": "published",
  ...
}
```

---

### Unpublish Form

**POST** `/api/forms/:id/unpublish` 🔒

Sets a form's status back to `draft`. Requires authentication.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "status": "draft",
  ...
}
```

---

### Delete Form

**DELETE** `/api/forms/:id` 🔒

Deletes a form and all its responses. Requires authentication.

**Response:** `204 No Content`

---

## Responses API

### Create Response

**POST** `/api/forms/:form_id/responses`

Submits a response to a form. No authentication required (public endpoint).

**Request Body:**

```json
{
  "data": {
    "question1": "answer1",
    "question2": "answer2"
  },
  "meta": {
    "ip": "192.168.1.1",
    "userAgent": "Mozilla/5.0..."
  }
}
```

**Response:** `201 Created`

```json
{
  "id": 1,
  "form_id": 1,
  "data": {
    "question1": "answer1",
    "question2": "answer2"
  },
  "meta": {
    "ip": "192.168.1.1",
    "userAgent": "Mozilla/5.0..."
  },
  "created_at": "2024-01-28T12:00:00Z"
}
```

---

### Get Response

**GET** `/api/responses/:id` 🔒

Retrieves a specific response by ID. Requires authentication.

**Response:** `200 OK`

```json
{
  "id": 1,
  "form_id": 1,
  "data": {...},
  "meta": {...},
  "created_at": "2024-01-28T12:00:00Z"
}
```

---

### Get Form Responses

**GET** `/api/forms/:id/responses` 🔒

Retrieves all responses for a form. Requires authentication.

**Response:** `200 OK`

```json
{
  "responses": [
    {
      "id": 1,
      "form_id": 1,
      "data": {...},
      "meta": {...},
      "created_at": "2024-01-28T12:00:00Z"
    }
  ],
  "count": 1
}
```

---

### Delete Response

**DELETE** `/api/responses/:id` 🔒

Deletes a specific response. Requires authentication.

**Response:** `204 No Content`

---

## Health Check Endpoints

### Basic Health Check

**GET** `/health`

Returns server status.

**Response:** `200 OK`

```json
{
  "status": "ok"
}
```

---

### Database Health Check

**GET** `/health/db`

Checks database connectivity.

**Response:** `200 OK`

```json
{
  "status": "ok",
  "database": "connected"
}
```

---

## Architecture

### Handler (`internal/form/handler.go`)

HTTP handlers that parse requests and return JSON responses.

| Handler         | Route                         | Auth | Description       |
| --------------- | ----------------------------- | ---- | ----------------- |
| `CreateForm`    | POST /api/forms               | 🔒   | Create a new form |
| `GetForm`       | GET /api/forms/:id            | 🔒   | Get form by ID    |
| `GetUserForms`  | GET /api/users/:id/forms      | 🔒   | Get user's forms  |
| `UpdateForm`    | PUT /api/forms/:id            | 🔒   | Update form       |
| `PublishForm`   | POST /api/forms/:id/publish   | 🔒   | Publish form      |
| `UnpublishForm` | POST /api/forms/:id/unpublish | 🔒   | Unpublish form    |
| `DeleteForm`    | DELETE /api/forms/:id         | 🔒   | Delete form       |

### Service (`internal/form/service.go`)

Business logic layer with validation and defaults.

| Method                  | Description                                        |
| ----------------------- | -------------------------------------------------- |
| `CreateForm`            | Creates form with default status, schema, settings |
| `GetFormByID`           | Retrieves form by ID                               |
| `GetFormByShareURL`     | Retrieves form by share URL                        |
| `GetUserForms`          | Gets all forms for a user                          |
| `GetUserPublishedForms` | Gets published forms only                          |
| `UpdateForm`            | Updates form fields                                |
| `PublishForm`           | Sets status to published                           |
| `UnpublishForm`         | Sets status to draft                               |
| `SetShareURL`           | Sets unique share URL                              |
| `DeleteForm`            | Deletes form                                       |

### Repository (`internal/form/repository.go`)

Data access layer using sqlc-generated queries.

| Method                 | SQL Query                    |
| ---------------------- | ---------------------------- |
| `Create`               | `CreateForm`                 |
| `GetByID`              | `GetFormByID`                |
| `GetByShareURL`        | `GetFormByShareURL`          |
| `GetByUserID`          | `ListFormsByUserID`          |
| `GetPublishedByUserID` | `ListPublishedFormsByUserID` |
| `Update`               | `UpdateForm`                 |
| `UpdateStatus`         | `UpdateFormStatus`           |
| `UpdateShareURL`       | `UpdateFormShareURL`         |
| `Delete`               | `DeleteForm`                 |

---

## Project Structure

```
internal/
  ├── config/          - Configuration management
  ├── database/        - Database connection and migrations
  │   ├── migrations/  - SQL migration files
  │   ├── queries/     - SQL query files for sqlc
  │   └── schema/      - Database schema definitions
  ├── db/              - Generated sqlc code
  ├── form/            - Form domain (handler, service, repository, model)
  ├── user/            - User domain (handler, service, repository, model)
  │   └── oauth/       - OAuth providers (Google)
  ├── response/        - Response domain (handler, service, repository, model)
  ├── middleware/      - Auth middleware
  └── shared/          - Shared utilities and helpers
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:

- `400 Bad Request` - Invalid request body or parameters
- `401 Unauthorized` - Missing or invalid authentication token
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## Notes

- 🔒 indicates endpoints that require JWT authentication
- All timestamps are in ISO 8601 format (UTC)
- JSON fields with `null` values may be omitted from responses
- `schema` field defaults to `[]` if not provided
- `settings` field defaults to `{}` if not provided
- Forms default to `draft` status when created
