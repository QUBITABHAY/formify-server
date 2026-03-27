# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project follows Semantic Versioning.

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
