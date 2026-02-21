-- Add Google Sheets integration columns to forms table
ALTER TABLE forms ADD COLUMN google_sheet_id TEXT;
ALTER TABLE forms ADD COLUMN google_sheet_name TEXT;
ALTER TABLE forms ADD COLUMN google_sheet_linked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE forms ADD COLUMN google_sheet_auto_sync BOOLEAN DEFAULT false;
