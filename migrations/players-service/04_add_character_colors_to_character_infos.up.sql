BEGIN;

ALTER TABLE character_infos
    ADD COLUMN IF NOT EXISTS skin_color JSONB NOT NULL DEFAULT '{"h":0,"s":0,"v":100}'::jsonb,
    ADD COLUMN IF NOT EXISTS hair_color JSONB NOT NULL DEFAULT '{"h":0,"s":0,"v":100}'::jsonb,
    ADD COLUMN IF NOT EXISTS eye_color JSONB NOT NULL DEFAULT '{"h":0,"s":0,"v":100}'::jsonb;

COMMIT;
