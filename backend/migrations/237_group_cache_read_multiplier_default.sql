-- Keep direct database writes aligned with the application-level default.
ALTER TABLE groups
    ALTER COLUMN cache_read_multiplier SET DEFAULT 1.1;
