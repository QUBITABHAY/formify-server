# API Documentation

Documentation for the Formify Server API endpoints.

---

## Base URL

The API runs on port `1323` by default. Configure via `PORT` environment variable.

```
http://localhost:1323
```

---

## Forms API

### Create Form

**POST** `/api/forms`

Creates a new form with default draft status.

**Request Body:**

```json
{
  "name": "Customer Survey",
  "description": "Optional description",
  "user_id": 1
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
  "share_url": null,
  "created_at": "2024-01-28T12:00:00Z",
  "updated_at": "2024-01-28T12:00:00Z"
}
```

---

### Get Form

**GET** `/api/forms/:id`

Retrieves a form by ID.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Customer Survey",
  "description": "Optional description",
  "user_id": 1,
  "status": "draft",
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

**GET** `/api/users/:id/forms`

Retrieves all forms belonging to a user.

**Response:** `200 OK`

```json
[
  {
    "id": 1,
    "name": "Customer Survey",
    "user_id": 1,
    "status": "draft",
    ...
  }
]
```

---

### Publish Form

**POST** `/api/forms/:id/publish`

Sets a form's status to `published`.

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

**POST** `/api/forms/:id/unpublish`

Sets a form's status back to `draft`.

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

**DELETE** `/api/forms/:id`

Deletes a form and all its responses.

**Response:** `204 No Content`

---

## Architecture

### Handler (`internal/handlers/form_handler.go`)

HTTP handlers that parse requests and return JSON responses.

| Handler         | Route                         | Description       |
| --------------- | ----------------------------- | ----------------- |
| `CreateForm`    | POST /api/forms               | Create a new form |
| `GetForm`       | GET /api/forms/:id            | Get form by ID    |
| `GetUserForms`  | GET /api/users/:id/forms      | Get user's forms  |
| `PublishForm`   | POST /api/forms/:id/publish   | Publish form      |
| `UnpublishForm` | POST /api/forms/:id/unpublish | Unpublish form    |
| `DeleteForm`    | DELETE /api/forms/:id         | Delete form       |

### Service (`internal/service/form_service.go`)

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
| `CountUserForms`        | Returns form count for user                        |

### Repository (`internal/repository/form_repository.go`)

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
| `CountByUserID`        | `CountFormsByUserID`         |
