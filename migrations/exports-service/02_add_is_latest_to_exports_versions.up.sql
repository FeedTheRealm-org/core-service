BEGIN;

ALTER TABLE exports_versions
ADD COLUMN IF NOT EXISTS is_latest BOOLEAN NOT NULL DEFAULT FALSE;

WITH latest_rows AS (
    SELECT DISTINCT ON (app_name, os) id
    FROM exports_versions
    ORDER BY app_name, os, created_at DESC, id DESC
)
UPDATE exports_versions
SET is_latest = TRUE
WHERE id IN (SELECT id FROM latest_rows);

COMMIT;
