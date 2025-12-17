BEGIN;

CREATE TABLE IF NOT EXISTS category_sprites (
    user_id UUID NOT NULL,
    category_id UUID NOT NULL,
    sprite_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, category_id)
);

COMMIT;
