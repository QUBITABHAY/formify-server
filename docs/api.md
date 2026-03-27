# API Documentation

Formify Server API reference.

## Base URL

Local default (from `make run`):

```text
http://localhost:1323
```

Docker Compose default:

```text
http://localhost:8080
```

## Authentication

Protected routes accept either:

- `Authorization: Bearer <jwt>`
- `token` cookie (HTTP-only, set by Google OAuth callback)

## Route Map

### Public

- `GET /` - basic service check
- `GET /health` - API health check
- `GET /health/db` - DB health check
- `POST /api/users` - create user
- `GET /api/forms/share/:share_url` - get published form
- `POST /api/forms/:form_id/responses` - submit response
- `POST /api/forms/:form_id/upload` - upload file for response usage
- `POST /api/auth/logout` - clear auth cookie
- `GET /api/auth/google` - start OAuth flow
- `GET /api/auth/google/callback` - finish OAuth flow and redirect

### Protected

- `GET /api/auth/me` - current authenticated user
- `GET /api/users/:id` - user by ID (self only)
- `GET /api/users/:id/forms` - forms by user (self only)
- `POST /api/forms` - create form
- `GET /api/forms/:id` - get form (owner only)
- `PUT /api/forms/:id` - update form (owner only)
- `DELETE /api/forms/:id` - delete form (owner only)
- `POST /api/forms/:id/publish` - publish form
- `POST /api/forms/:id/unpublish` - unpublish form
- `POST /api/forms/:id/sheets/create` - create and link Google Sheet
- `DELETE /api/forms/:id/sheets/link` - unlink Google Sheet
- `GET /api/forms/:id/responses` - list responses for form
- `GET /api/responses/:id` - get response by ID
- `DELETE /api/responses/:id` - delete response

## Request and Response Contracts

### Create User

`POST /api/users`

Request:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "strong-password"
}
```

Success `201`:

```json
{
  "id": 1,
  "name": "Jane Doe",
  "email": "jane@example.com"
}
```

### Auth Me

`GET /api/auth/me` (protected)

Success `200`:

```json
{
  "user": {
    "id": 1,
    "name": "Jane Doe",
    "email": "jane@example.com"
  }
}
```

### Create Form

`POST /api/forms` (protected)

Request:

```json
{
  "name": "Customer Survey",
  "description": "Q2 feedback form",
  "schema": [],
  "settings": {}
}
```

Success `201`:

```json
{
  "id": 1,
  "name": "Customer Survey",
  "description": "Q2 feedback form",
  "user_id": 1,
  "status": "draft",
  "schema": [],
  "settings": {},
  "share_url": null,
  "google_sheet_auto_sync": false,
  "created_at": "2026-03-27T10:00:00Z",
  "updated_at": "2026-03-27T10:00:00Z"
}
```

### Submit Response

`POST /api/forms/:form_id/responses`

Request:

```json
{
  "data": {
    "q1": "Great experience"
  },
  "meta": {
    "source": "web"
  }
}
```

Success `201`:

```json
{
  "id": 10,
  "form_id": 1,
  "data": {
    "q1": "Great experience"
  },
  "meta": {
    "source": "web"
  },
  "created_at": "2026-03-27T10:05:00Z"
}
```

### List Form Responses

`GET /api/forms/:id/responses` (protected)

Success `200`:

```json
{
  "form_id": 1,
  "count": 1,
  "responses": [
    {
      "id": 10,
      "form_id": 1,
      "data": {
        "q1": "Great experience"
      },
      "meta": {
        "source": "web"
      },
      "created_at": "2026-03-27T10:05:00Z"
    }
  ]
}
```

### Upload File

`POST /api/forms/:form_id/upload`

Request content type: `multipart/form-data`

Required field:

- `file` (max 10 MB)

Allowed file types:

- `image/jpeg`
- `image/png`
- `image/gif`
- `image/webp`
- `application/pdf`
- `application/zip`

Success `200`:

```json
{
  "public_id": "formify/1/abc123",
  "url": "https://res.cloudinary.com/...",
  "format": "png",
  "bytes": 24567
}
```

### Create and Link Google Sheet

`POST /api/forms/:id/sheets/create` (protected)

Request (optional body):

```json
{
  "title": "Survey Responses"
}
```

Success `201`:

```json
{
  "form": {
    "id": 1,
    "name": "Customer Survey",
    "google_sheet_id": "spreadsheet_id",
    "google_sheet_name": "Survey Responses",
    "google_sheet_auto_sync": true
  },
  "spreadsheet_id": "spreadsheet_id",
  "spreadsheet_url": "https://docs.google.com/spreadsheets/d/spreadsheet_id"
}
```

## OAuth Flow

1. Call `GET /api/auth/google`.
2. Complete Google sign-in and consent.
3. Backend handles `GET /api/auth/google/callback`.
4. Backend sets `token` cookie and redirects to:
   `FRONTEND_URL/auth/callback`

## Error Format

Errors use JSON with an `error` field.

```json
{
  "error": "Form not found"
}
```

Common statuses:

- `400` bad input or invalid path params
- `401` unauthorized or missing Google OAuth token (for Sheets)
- `403` access denied or form not accepting submissions
- `404` resource not found
- `500` internal server error
