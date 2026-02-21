-- Remove Google Sheets integration columns from forms table
ALTER TABLE forms DROP COLUMN IF EXISTS google_sheet_id;
ALTER TABLE forms DROP COLUMN IF EXISTS google_sheet_name;
ALTER TABLE forms DROP COLUMN IF EXISTS google_sheet_linked_at;
ALTER TABLE forms DROP COLUMN IF EXISTS google_sheet_auto_sync;
