BEGIN;

CREATE TABLE IF NOT EXISTS exports_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name TEXT NOT NULL,
    version TEXT NOT NULL,
    os TEXT NOT NULL,
    path TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_name, version, os)
);

COMMIT;
