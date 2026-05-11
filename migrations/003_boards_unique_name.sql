-- Add unique constraint on board name to prevent duplicates
ALTER TABLE boards ADD CONSTRAINT boards_name_unique UNIQUE (name);
