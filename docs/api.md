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

- `POST /api/users` - Create user account
- `GET /api/forms/share/:share_url` - Get published form by share URL
- `POST /api/forms/:form_id/responses` - Submit form response
- `GET /api/auth/google` - Google OAuth login
- `GET /api/auth/google/callback` - Google OAuth callback

---

## Authentication API

### Google OAuth Login

**GET** `/api/auth/google`

Initiates Google OAuth flow. Redirects to Google login page.

---

### Google OAuth Callback

**GET** `/api/auth/google/callback`

Handles OAuth callback from Google.

Behavior:

- Creates or links user account
- Stores/refreshes Google OAuth tokens for Sheets usage
- Sets JWT as HTTP-only `token` cookie
- Redirects to frontend callback URL

**Response:** `307 Temporary Redirect`

**Redirect Target:** `{FRONTEND_URL}/auth/callback`

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


### Get Public Form

**GET** `/api/forms/share/:share_url`

Retrieves a published form by share URL. No authentication required.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "description": "Optional description",
  "user_id": 1,
  "status": "published",
  "schema": [],
  "settings": {},
  "share_url": "AbCdEf123...",
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
If the form does not already have a `share_url`, one is generated.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "status": "published",
  "share_url": "AbCdEf123...",
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
Only published forms accept responses.

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

**Error:** `403 Forbidden`

```json
{
  "error": "Form is not accepting responses"
}
```

**Error:** `404 Not Found`

```json
{
  "error": "Form not found"
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

## Google Sheets Integration API

### Create and Link Google Sheet

**POST** `/api/forms/:id/sheets/create` 🔒

Creates a new Google Sheet and links it to the form using the authenticated user's Google OAuth token. Form responses are exported with appropriate column headers. Requires authentication.

**Request Body:**

```json
{
  "title": "Optional Custom Title"
}
```

If `title` is omitted, defaults to `"Form Name - Responses"`.

**Response:** `201 Created`

```json
{
  "form": {
    "id": 1,
    "name": "Customer Survey",
    "google_sheet_id": "GENERATED_ID",
    "google_sheet_name": "Customer Survey - Responses",
    "google_sheet_auto_sync": true,
    ...
  },
  "spreadsheet_id": "GENERATED_ID",
  "spreadsheet_url": "https://docs.google.com/spreadsheets/d/GENERATED_ID"
}
```

**Note:** The new spreadsheet will have "Form Responses" sheet with headers: "Submission ID", "Submitted At", followed by form field names.

---

### Unlink Google Sheet

**DELETE** `/api/forms/:id/sheets/link` 🔒

Removes the Google Sheet link from a form. Requires authentication.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "google_sheet_id": null,
  "google_sheet_name": null,
  "google_sheet_auto_sync": false,
  ...
}
```

---

### Sheets Auth Strategy

Sheets operations require the form owner's Google OAuth access token.

If the user does not have a valid OAuth token, sheets operations fail and the user must log in with Google again.

---

### Auto-Sync Behavior

When a form response is submitted and `google_sheet_auto_sync` is enabled:

1. Response is saved to database (synchronously)
2. Background task appends response to the linked Google Sheet (asynchronously)
3. Response data is converted to spreadsheet row format matching form schema
4. Any sync errors are logged but do not affect response creation

**Note:** Auto-sync errors do not cause the response submission to fail. Check server logs for sync issues.

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
| `GetPublicFormsByShareURL` | GET /api/forms/share/:share_url |      | Get form by share URL |
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
