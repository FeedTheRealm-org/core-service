BEGIN;

-- Make 'category' column nullable on item_sprites so application which writes only category_id works
ALTER TABLE item_sprites
    ALTER COLUMN category DROP NOT NULL;

COMMIT;
