BEGIN;

ALTER TABLE character_infos
    DROP COLUMN IF EXISTS skin_color,
    DROP COLUMN IF EXISTS hair_color,
    DROP COLUMN IF EXISTS eye_color;

COMMIT;
