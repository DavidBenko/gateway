ALTER TABLE environments ADD COLUMN session_type TEXT NOT NULL DEFAULT '';
ALTER TABLE environments ADD COLUMN session_header TEXT NOT NULL DEFAULT '';
UPDATE environments SET session_type='client';
