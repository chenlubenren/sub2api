ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS night_cache_read_multiplier NUMERIC(10,4) NOT NULL DEFAULT 1.2;

-- The existing field is the regular-hours multiplier. Preserve explicitly
-- configured values while normalizing the old default of 2.0 to 1.1 only for
-- rows that still carry the migration default.
UPDATE groups
SET cache_read_multiplier = 1.1
WHERE cache_read_multiplier = 2.0;
