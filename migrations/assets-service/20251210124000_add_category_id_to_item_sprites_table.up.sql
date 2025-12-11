BEGIN;

ALTER TABLE item_sprites ADD COLUMN IF NOT EXISTS category_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_item_sprites_category'
    ) THEN
        ALTER TABLE item_sprites
            ADD CONSTRAINT fk_item_sprites_category FOREIGN KEY (category_id) REFERENCES item_categories(id) ON DELETE RESTRICT;
    END IF;
END;
$$;

COMMIT;
