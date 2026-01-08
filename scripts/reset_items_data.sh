#!/usr/bin/env bash
set -euo pipefail

# Reset items and their sprites via HTTP API (development only)
# - Fetches ALL items from /items/metadata
# - Deletes each referenced item sprite via /assets/sprites/items/{sprite_id}
# - Deletes each item via /items/{id}
# - Optionally deletes local files under ./bucket/sprites/items

USAGE="Usage: $0 [--base-url <http://host:port>] [--delete-files] [-y]\n
Options:\n  --base-url <url>  Base URL of core-service API (default: http://localhost:8000)\n  --delete-files    Also delete the files under ./bucket/sprites/items (prompt required)\n  -y                Skip confirmation prompt (dangerous)\n"

BASE_URL="http://localhost:8000"
DELETE_FILES=false
SKIP_PROMPT=false

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --base-url)
      BASE_URL="$2"
      shift 2
      ;;
    --delete-files)
      DELETE_FILES=true
      shift
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

if ! command -v curl >/dev/null 2>&1; then
  echo "Error: curl is required to run this script" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required to run this script" >&2
  exit 1
fi

if [ "$SKIP_PROMPT" = false ]; then
  echo "This will delete ALL items and their associated sprites using the HTTP API at: $BASE_URL."
  if [ "$DELETE_FILES" = true ]; then
    echo "It will also DELETE files under './bucket/sprites/items' on disk."
  fi
  read -p "Type 'YES' to continue: " CONFIRM
  if [ "$CONFIRM" != "YES" ]; then
    echo "Aborted by user"
    exit 1
  fi
fi

echo "Fetching items from $BASE_URL/items/metadata ..."
ITEMS_JSON=$(curl -fsS "$BASE_URL/items/metadata")

# Extract item IDs and sprite IDs from response (wrapped in .data)
ITEM_IDS=$(echo "$ITEMS_JSON" | jq -r '.data.items[]?.id' 2>/dev/null || true)
SPRITE_IDS=$(echo "$ITEMS_JSON" | jq -r '.data.items[]?.sprite_id' 2>/dev/null | grep -v '^00000000-0000-0000-0000-000000000000$' | sort -u || true)

if [ -z "${ITEM_IDS:-}" ]; then
  echo "No items found. Nothing to delete from API."
else
  if [ -n "${SPRITE_IDS:-}" ]; then
    echo "Deleting sprites referenced by items..."
    while IFS= read -r SPRITE_ID; do
      [ -z "$SPRITE_ID" ] && continue
      echo "- Deleting sprite $SPRITE_ID ..."
      if ! curl -fsS -X DELETE "$BASE_URL/assets/sprites/items/$SPRITE_ID" >/dev/null; then
        echo "  Warning: failed to delete sprite $SPRITE_ID (might not exist)" >&2
      fi
    done <<< "$SPRITE_IDS"
  else
    echo "No sprite IDs referenced by items. Skipping sprite deletion."
  fi

  echo "Deleting items..."
  while IFS= read -r ITEM_ID; do
    [ -z "$ITEM_ID" ] && continue
    echo "- Deleting item $ITEM_ID ..."
    if ! curl -fsS -X DELETE "$BASE_URL/items/$ITEM_ID" >/dev/null; then
      echo "  Warning: failed to delete item $ITEM_ID (might not exist)" >&2
    fi
  done <<< "$ITEM_IDS"
fi

echo "Fetching all item sprites from $BASE_URL/assets/sprites/items ..."
ALL_SPRITES_JSON=$(curl -fsS "$BASE_URL/assets/sprites/items")
REMAINING_SPRITE_IDS=$(echo "$ALL_SPRITES_JSON" | jq -r '.data.sprites[]?.id' 2>/dev/null || true)

if [ -z "${REMAINING_SPRITE_IDS:-}" ]; then
  echo "No remaining sprites found (or already deleted)."
else
  echo "Deleting remaining sprites (orphans or already removed in previous step)..."
  while IFS= read -r SPRITE_ID; do
    [ -z "$SPRITE_ID" ] && continue
    echo "- Deleting sprite $SPRITE_ID ..."
    if ! curl -fsS -X DELETE "$BASE_URL/assets/sprites/items/$SPRITE_ID" >/dev/null; then
      echo "  Warning: failed to delete sprite $SPRITE_ID (might not exist)" >&2
    fi
  done <<< "$REMAINING_SPRITE_IDS"
fi

if [ "$DELETE_FILES" = true ]; then
  if [ "$SKIP_PROMPT" = false ]; then
    read -p "Also delete files under ./bucket/sprites/items? Type 'YES' to continue: " CONFIRM2
    if [ "$CONFIRM2" != "YES" ]; then
      echo "Skipping file deletion"
      echo "Reset via API completed"
      exit 0
    fi
  fi
  echo "Deleting files under ./bucket/sprites/items..."
  rm -rf ./bucket/sprites/items/* || true
  echo "Deleted files under ./bucket/sprites/items"
fi

echo "Reset via API completed"
