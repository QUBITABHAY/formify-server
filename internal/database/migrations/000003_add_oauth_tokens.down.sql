-- Remove OAuth token columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS google_access_token;
ALTER TABLE users DROP COLUMN IF EXISTS google_refresh_token;
ALTER TABLE users DROP COLUMN IF EXISTS google_token_expiry;
