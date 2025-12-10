BEGIN;

ALTER TABLE item_sprites DROP CONSTRAINT IF EXISTS fk_item_sprites_category;
ALTER TABLE item_sprites DROP COLUMN IF EXISTS category_id;

COMMIT;
