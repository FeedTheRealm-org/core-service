#!/usr/bin/env bash
set -euo pipefail

# Reset items, item_sprites and item_categories data in the DB (development only)
# Also optionally deletes files under ./bucket/sprites/items

USAGE="Usage: $0 [--docker] [--delete-files] [--db-url <postgres://...>]\n
Options:\n  --docker         Run DB commands inside docker-compose dev_db container (default)\n  --no-docker      Use local psql and pass DB URL via --db-url\n  --delete-files   Also delete the files under ./bucket/sprites/items (prompt required)\n  --db-url <url>   When using --no-docker, provide the DATABASE_URL for psql\n  -y               Skip confirmation prompt (dangerous)\n"

USE_DOCKER=true
DELETE_FILES=false
DB_URL=""
SKIP_PROMPT=false

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --docker)
      USE_DOCKER=true
      shift
      ;;
    --no-docker)
      USE_DOCKER=false
      shift
      ;;
    --delete-files)
      DELETE_FILES=true
      shift
      ;;
    --db-url)
      DB_URL="$2"
      shift 2
      ;;
    -y)
      SKIP_PROMPT=true
      shift
      ;;
    -h|--help)
      echo -e "$USAGE"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo -e "$USAGE"
      exit 1
      ;;
  esac
done

SQL=$(cat <<'SQL'
BEGIN;
DELETE FROM items;
DELETE FROM item_sprites;
DELETE FROM item_categories;
COMMIT;
SQL
)

if [ "$SKIP_PROMPT" = false ]; then
  echo "This will delete ALL records in tables: items, item_sprites, item_categories."
  if [ "$DELETE_FILES" = true ]; then
    echo "It will also DELETE files under './bucket/sprites/items' on disk."
  fi
  read -p "Type 'YES' to continue: " CONFIRM
  if [ "$CONFIRM" != "YES" ]; then
    echo "Aborted by user"
    exit 1
  fi
fi

if [ "$USE_DOCKER" = true ]; then
  echo "Running reset against dev_db container via docker-compose..."
  docker compose -f docker-compose.dev.yml exec -T dev_db psql -v ON_ERROR_STOP=1 -U postgres -d postgres -c "$SQL"
else
  if [ -z "$DB_URL" ]; then
    echo "Error: --db-url is required when using --no-docker"
    exit 1
  fi
  echo "Running reset against DB: $DB_URL"
  # Use psql with provided DB_URL
  psql "$DB_URL" -v ON_ERROR_STOP=1 -c "$SQL"
fi

if [ "$DELETE_FILES" = true ]; then
  if [ "$SKIP_PROMPT" = false ]; then
    read -p "Also delete files under ./bucket/sprites/items? Type 'YES' to continue: " CONFIRM2
    if [ "$CONFIRM2" != "YES" ]; then
      echo "Skipping file deletion"
      exit 0
    fi
  fi
  echo "Deleting files under ./bucket/sprites/items..."
  rm -rf ./bucket/sprites/items/* || true
  echo "Deleted files under ./bucket/sprites/items"
fi

echo "Reset completed"
