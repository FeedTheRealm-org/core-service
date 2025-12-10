BEGIN;

-- Make 'category' column nullable on items so application which writes only category_id works
ALTER TABLE items
    ALTER COLUMN category DROP NOT NULL;

COMMIT;
