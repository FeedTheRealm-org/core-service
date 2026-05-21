BEGIN;

ALTER TABLE exports_versions
DROP COLUMN IF EXISTS is_latest;

COMMIT;
