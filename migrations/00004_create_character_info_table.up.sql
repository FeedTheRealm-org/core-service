BEGIN;

CREATE TABLE IF NOT EXISTS character_infos (
    user_id UUID PRIMARY KEY,
    character_name TEXT NOT NULL UNIQUE,
    character_bio TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
