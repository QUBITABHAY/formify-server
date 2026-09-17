# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project follows Semantic Versioning.

## [Release 0.0.2]

### Security

- **CSRF Cookie Hardening**: Configured authentication session cookies with `SameSite=Lax` mode unconditionally in `internal/auth/handler.go`, preventing cross-site cookie transmission and protecting mutating endpoints from Cross-Site Request Forgery (CSRF).
- **Google Sheets Formula Injection Sanitization**: Sanitized spreadsheet cell values in `internal/integrations/google` by prepending a single quote `'` to inputs beginning with formula triggers (`=`, `+`, `-`, `@`, `\t`, `\r`) during row formatting and `AppendRow`.
- **Public Form Metadata & Quiz Answer Key Stripping**: Purged `correctAnswer` and `correct_answer` fields recursively from form schemas returned by `GET /api/forms/share/:share_url` and omitted internal Google Sheet metadata to prevent answer key and integration data leakage to respondents.
- **Safe ZIP Archive Inspection**: Re-enabled `application/zip` support with in-memory archive inspection in `internal/file_upload`: rejects dangerous executable and script files (`.exe`, `.bat`, `.sh`, `.vbs`, etc.), blocks directory traversal (`../`), and mitigates decompression bombs (max 500 files, 50MB uncompressed limit).
- **Form Upload Schema Verification**: Enforced verification that target forms actively include a file upload field before permitting file uploads.

### Performance

- **Neon PostgreSQL Connection Pool Tuning**: Configured `pgxpool` with `MinConns = 3` to keep warm connections alive (eliminating 80–150ms cold-start TLS handshakes with Neon), `MaxConns = 20`, and explicit connection idle/lifetime bounds.

### Fixed

- **Async Sheets Synchronization Context**: Fixed goroutine cancellation in background Google Sheets sync by wrapping request contexts with `context.WithoutCancel`.

## [Released 0.0.1]

### Added

- Initial public backend API for users, forms, and responses.
- JWT-based authentication and Google OAuth login flow.
- Google Sheets integration for form response workflows.
- PostgreSQL schema, migrations, and sqlc-generated query layer.
- Deployment artifacts for Docker and Azure Container Instances.
- Cloudinary-based file upload flow for forms, including upload endpoint and validation.
- OAuth callback redirect improvements to support frontend callback routing.
- Package-level comments across internal modules for better codebase documentation.

### Changed

- Integrated structured logging with Zap across API, services, middleware, and migration command paths.
- Improved server bootstrap with clearer middleware and route setup.
- Refactored auth, form, response, and migration layers for cleaner function signatures and stronger type consistency.
- Improved error handling and operational logging in database init, migrations, and file upload execution paths.
- Replaced hardcoded auth provider values with shared constants.
- Updated Google provider scope usage and refined OAuth user/token handling.
- Refined Google Sheets integration logic and form schema parsing behavior.
- Removed legacy email/password login flow in favor of OAuth-centric auth flow updates.
- Updated Docker and repository hygiene (Dockerfile/.gitignore adjustments and deployment workflow cleanup).
- Expanded and refreshed project documentation (README, API docs, setup guide, logging guide).
