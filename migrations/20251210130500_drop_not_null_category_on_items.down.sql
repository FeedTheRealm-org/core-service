BEGIN;

-- Restore NOT NULL to category column on items
ALTER TABLE items
    ALTER COLUMN category SET NOT NULL;

COMMIT;
