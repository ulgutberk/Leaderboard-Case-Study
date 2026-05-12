ALTER TABLE boards ADD COLUMN IF NOT EXISTS last_reset_at TIMESTAMPTZ;

UPDATE boards
SET last_reset_at = created_at
WHERE last_reset_at IS NULL;

ALTER TABLE boards ALTER COLUMN last_reset_at SET DEFAULT NOW();
ALTER TABLE boards ALTER COLUMN last_reset_at SET NOT NULL;