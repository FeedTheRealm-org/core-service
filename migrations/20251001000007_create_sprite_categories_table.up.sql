BEGIN;
CREATE TABLE IF NOT EXISTS sprite_categories (
    sprite_id UUID NOT NULL REFERENCES sprites(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (sprite_id, category_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sprite_categories_sprite_id ON sprite_categories(sprite_id);
CREATE INDEX IF NOT EXISTS idx_sprite_categories_category_id ON sprite_categories(category_id);
COMMIT;
