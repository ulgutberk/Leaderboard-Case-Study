-- Add created_at to scores to track first submission time.
-- Used for tie-breaking: equal scores → earlier submission ranks higher.
ALTER TABLE scores ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
