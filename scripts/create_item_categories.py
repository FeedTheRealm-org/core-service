#!/usr/bin/python3

import requests
import sys
from typing import List, Dict, Optional


def fetch_item_categories(server_url: str) -> Optional[List[Dict]]:
    url = f"{server_url}/items/categories"
    try:
        resp = requests.get(url)
        if resp.status_code != 200:
            print(f"Failed to fetch item categories: {resp.status_code}")
            print(resp.text)
            return None
        data = resp.json().get("data", {})
        categories = data.get("categories", [])

        print("\n=== Current Item Categories ===")
        if categories:
            for idx, cat in enumerate(categories, 1):
                print(f"{idx}. {cat['name']} (ID: {cat['id']})")
        else:
            print("No item categories found.")
        print()
        return categories
    except Exception as e:
        print(f"Error fetching item categories: {e}")
        return None


def add_item_category(server_url: str, name: str) -> Optional[Dict]:
    url = f"{server_url}/items/categories"
    payload = {"name": name}
    try:
        resp = requests.post(url, json=payload)
        if resp.status_code in (200, 201):
            cat = resp.json().get("data", {})
            print(f"✓ Created item category: {cat.get('name')} (ID: {cat.get('id')})")
            return cat
        elif resp.status_code == 409:
            print(f"✗ Item category '{name}' already exists")
            return None
        else:
            print(f"✗ Failed to create item category '{name}': {resp.status_code}")
            print(resp.text)
            return None
    except Exception as e:
        print(f"✗ Error creating item category '{name}': {e}")
        return None


def seed_default_item_categories(server_url: str) -> List[Dict]:
    print("Seeding default item categories (you can edit these later)...")
    default_names = [
        "Weapons",
        "Armor",
        "Potions",
        #"Scrolls",
        #"Materials",
    ]

    existing = fetch_item_categories(server_url) or []
    existing_by_name = {c["name"].lower(): c for c in existing}

    created_or_existing: List[Dict] = []

    for name in default_names:
        if name.lower() in existing_by_name:
            cat = existing_by_name[name.lower()]
            print(f"= Exists item category: {cat['name']} (ID: {cat['id']})")
            created_or_existing.append(cat)
        else:
            cat = add_item_category(server_url, name)
            if cat:
                created_or_existing.append(cat)

    print()
    return created_or_existing


def main() -> None:
    if len(sys.argv) < 2:
        print("Usage: python create_item_categories.py <server_url>")
        print("Example: python create_item_categories.py http://localhost:8000")
        return

    server_url = sys.argv[1].rstrip("/")
    # Ensure scheme
    if not server_url.startswith("http://") and not server_url.startswith("https://"):
        server_url = "http://" + server_url

    seed_default_item_categories(server_url)


if __name__ == "__main__":
    main()
