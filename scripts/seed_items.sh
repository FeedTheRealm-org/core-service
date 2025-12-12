#!/bin/bash
set -euo pipefail

# Seed items and their sprites using 6000FantasyIcons from the Unity client.
# Usage: ./scripts/seed_items.sh <FTR_CLIENT_PATH> [SERVER_URL]
# Example: ./scripts/seed_items.sh ../client http://localhost:8000

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
    echo "Usage: $0 <FTR_CLIENT_PATH> [SERVER_URL]"
    exit 1
fi

FTR_CLIENT_PATH=$1
SERVER_URL=${2:-http://localhost:8000}

# Ensure SERVER_URL includes the http/https scheme otherwise requests will fail
if [[ ! "$SERVER_URL" =~ ^https?:// ]]; then
    SERVER_URL="http://$SERVER_URL"
fi

ICONS_PATH="$FTR_CLIENT_PATH/Assets/6000FantasyIcons"

if [ ! -d "$ICONS_PATH" ]; then
    echo "Error: icons folder not found at '$ICONS_PATH'"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "Seeding 5 Weapons and 5 Armor items with sprites from $ICONS_PATH ..."
python3 "$SCRIPT_DIR/seed_items_with_sprites.py" "$SERVER_URL" "$ICONS_PATH" 5 5

echo "Done seeding items and sprites."
