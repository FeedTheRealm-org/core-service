BEGIN;

CREATE TABLE IF NOT EXISTS materials (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    world_id UUID NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    PRIMARY KEY (id, world_id)
);

CREATE INDEX IF NOT EXISTS idx_materials_world_id ON materials(world_id);

COMMIT;
