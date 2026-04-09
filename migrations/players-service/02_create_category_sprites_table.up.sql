BEGIN;

CREATE TABLE IF NOT EXISTS category_sprites (
    user_id UUID NOT NULL,
    category_id UUID NOT NULL,
    sprite_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, category_id),
    CONSTRAINT fk_category_sprites_character_info
        FOREIGN KEY (user_id) REFERENCES character_infos(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_category_sprites_user_id ON category_sprites(user_id);

COMMIT;
