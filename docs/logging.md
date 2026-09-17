# Structured Logging with Zap

This project uses Zap for structured logging across API, migrations, services, and HTTP request middleware.

## Environment Variable

Logging mode is controlled by ENV.

- development: console output with colored levels
- production: JSON output for log collection

Examples:

```bash
ENV=development go run ./cmd/api
ENV=production go run ./cmd/api
```

## Initialization

Initialize logging at process startup:

```go
if err := logger.InitFromEnv(); err != nil {
    panic(err)
}
defer logger.Close()

log := logger.GetLogger()
```

## HTTP Request Logging

Request logging middleware is centralized in `internal/logger/middleware.go`.

Use it in API bootstrap:

```go
e.Use(logger.RequestLogger())
```

This logs request metadata (method, uri, status, latency, host, remote_ip, user_agent, request_id) using the same Zap output mode configured by ENV.

## Logger Helpers

Available helpers in `internal/logger/logger.go`:

- logger.Init(environment)
- logger.InitFromEnv()
- logger.GetLogger()
- logger.GetSugaredLogger()
- logger.Close()
- logger.ToField(key, value)
- logger.ToFields(values...)

## Usage Examples

Structured logger:

```go
log := logger.GetLogger()
log.Info("form created", zap.Int32("form_id", formID), zap.Int32("user_id", userID))
log.Warn("sheet sync skipped", zap.String("reason", "token_missing"))
log.Error("failed to append row", zap.Error(err))
```

Sugared logger:

```go
sugar := logger.GetSugaredLogger()
sugar.Infof("processing form %d", formID)
```

## Notes

- Keep logs structured with fields instead of string interpolation.
- Avoid logging secrets or full credentials.
- Prefer logging once at the layer with enough business context.
