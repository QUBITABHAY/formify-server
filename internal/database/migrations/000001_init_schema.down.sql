-- Drop indexes
DROP INDEX IF EXISTS idx_responses_form_id;
DROP INDEX IF EXISTS idx_forms_status;
DROP INDEX IF EXISTS idx_forms_user_id;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS responses;
DROP TABLE IF EXISTS forms;
DROP TABLE IF EXISTS users;

-- Drop enum
DROP TYPE IF EXISTS form_status;
