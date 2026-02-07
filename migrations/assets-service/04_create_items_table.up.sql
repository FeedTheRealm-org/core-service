BEGIN;

CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL,
    category_id UUID NOT NULL REFERENCES items_categories(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_items_world_id ON items(world_id);
CREATE INDEX IF NOT EXISTS idx_items_category_id ON items(category_id);

COMMIT;
