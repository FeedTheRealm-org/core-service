BEGIN;
DROP INDEX IF EXISTS idx_sprite_categories_category_id;
DROP INDEX IF EXISTS idx_sprite_categories_sprite_id;
DROP TABLE IF EXISTS sprite_categories;
COMMIT;
