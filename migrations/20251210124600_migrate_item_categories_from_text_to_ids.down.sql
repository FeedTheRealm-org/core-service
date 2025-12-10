BEGIN;

-- Data migration is not trivially reversible: this down migration is a no-op
-- If needed, restore from backups or add a migration to revert by name.

COMMIT;
