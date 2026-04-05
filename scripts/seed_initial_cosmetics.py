#!/usr/bin/env python3

# Get jwt token with:
# curl -X POST localhost:8000/auth/login -H "Content-Type: text/json" -d '{"email": "admin@admin.admin", "password": "admin123"}' | jq -r '.data.access_token'

import getpass
import os
import sys
from pathlib import Path

import requests


CATEGORY_MAPPINGS = [
    ("ArmorHelmet", "Assets/HeroEditor4D/FantasyHeroes/Sprites/Equipment/Armor/Basic"),
    ("ArmorBody", "Assets/HeroEditor4D/FantasyHeroes/Sprites/Equipment/Armor/Basic"),
    ("ArmorLegR", "Assets/HeroEditor4D/FantasyHeroes/Sprites/Equipment/Armor/Basic"),
    ("Hair", "Assets/HeroEditor4D/Common/Sprites/BodyParts/Hair/Basic"),
    ("Beard", "Assets/HeroEditor4D/Common/Sprites/BodyParts/Beard/Basic"),
    ("EyeBrows", "Assets/HeroEditor4D/Common/Sprites/BodyParts/Eyebrows/Basic"),
    ("Eyes", "Assets/HeroEditor4D/Common/Sprites/BodyParts/Eyes/Basic"),
    ("Mouth", "Assets/HeroEditor4D/Common/Sprites/BodyParts/Mouth/Basic"),
    ("EarringR", "Assets/HeroEditor4D/Common/Sprites/Equipment/Earrings/Common"),
    ("Back", "Assets/HeroEditor4D/Common/Sprites/Equipment/Back/Common"),
    ("Mask", "Assets/HeroEditor4D/Common/Sprites/Equipment/Mask/Common"),
]

SHARED_CATEGORY_SOURCES = {
    "ArmorBody": "ArmorHelmet",
    "ArmorLegR": "ArmorHelmet",
}


def usage() -> None:
    print("Usage: seed_categories_and_sprites.py <SERVER_URL> <SPRITES_BASE_PATH>")
    print("Reads token from JWT_TOKEN env var; if empty/unset, asks via stdin.")


def get_token() -> str:
    token = os.getenv("JWT_TOKEN", "")
    if token == "":
        token = getpass.getpass("JWT token (press Enter for none): ")
    return token


def build_headers(token: str, include_json: bool = False) -> dict:
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if include_json:
        headers["Content-Type"] = "application/json"
    return headers


def fetch_categories(server_url: str, token: str) -> dict:
    url = f"{server_url}/assets/cosmetics/categories"
    try:
        response = requests.get(url, headers=build_headers(token), timeout=30)
    except requests.RequestException as exc:
        print(f"Failed to fetch categories: {exc}")
        return {}

    if response.status_code != 200:
        print(
            f"Failed to fetch categories: HTTP {response.status_code} - {response.text}")
        return {}

    try:
        payload = response.json()
    except ValueError:
        print("Failed to parse categories response JSON")
        return {}

    category_list = payload.get("data", {}).get("category_list", [])
    return {
        item.get("category_name"): item.get("category_id")
        for item in category_list
        if item.get("category_name") is not None and item.get("category_id") is not None
    }


def create_category(server_url: str, category_name: str, token: str):
    url = f"{server_url}/assets/cosmetics/categories"
    payload = {"category_name": category_name}

    try:
        response = requests.post(
            url,
            json=payload,
            headers=build_headers(token),
            timeout=30,
        )
    except requests.RequestException as exc:
        print(f"  Failed creating category '{category_name}': {exc}")
        return None

    if response.status_code in (200, 201):
        try:
            data = response.json().get("data", {})
            return data.get("category_id")
        except ValueError:
            return None

    if response.status_code == 409:
        return "CONFLICT"

    print(
        f"  Failed creating category '{category_name}': HTTP {response.status_code} - {response.text}")
    return None


def fetch_cosmetics_by_category(server_url: str, category_id, token: str) -> list:
    url = f"{server_url}/assets/cosmetics/categories/{category_id}"
    try:
        response = requests.get(url, headers=build_headers(token), timeout=30)
    except requests.RequestException as exc:
        print(f"  Failed fetching cosmetics for category {category_id}: {exc}")
        return []

    if response.status_code != 200:
        print(
            f"  Failed fetching cosmetics for category {category_id}: HTTP {response.status_code} - {response.text}")
        return []

    try:
        payload = response.json()
    except ValueError:
        print(f"  Failed parsing cosmetics JSON for category {category_id}")
        return []

    return payload.get("data", {}).get("cosmetics_list", [])


def upload_sprite(server_url: str, category_id, sprite_path: Path, token: str):
    url = f"{server_url}/assets/cosmetics/categories/{category_id}"
    headers = build_headers(token)

    with sprite_path.open("rb") as sprite_file:
        files = {"sprite": (sprite_path.name, sprite_file, "image/png")}
        data = {"category_id": str(category_id)}
        try:
            response = requests.put(
                url, files=files, data=data, headers=headers, timeout=60)
        except requests.RequestException as exc:
            return False, str(exc)

    if response.status_code in (200, 201):
        return True, ""
    return False, f"HTTP {response.status_code} - {response.text}"


def link_sprite_by_id(server_url: str, category_id, sprite_id, token: str):
    url = f"{server_url}/assets/cosmetics/categories/{category_id}/sprites/{sprite_id}"

    try:
        response = requests.put(url, headers=build_headers(token), timeout=30)
    except requests.RequestException as exc:
        return False, str(exc)

    if response.status_code in (200, 201):
        return True, ""
    return False, f"HTTP {response.status_code} - {response.text}"


def main() -> int:
    if len(sys.argv) != 3:
        usage()
        return 1

    server_url = sys.argv[1].rstrip("/")
    sprites_base_path = Path(sys.argv[2])

    if not sprites_base_path.exists() or not sprites_base_path.is_dir():
        print(f"Error: sprites base path not found: {sprites_base_path}")
        return 1

    token = get_token()

    print("Seeding categories and uploading sprites...")

    category_ids = fetch_categories(server_url, token)

    created_categories = 0
    uploaded_sprites = 0
    linked_sprites = 0
    failed_uploads = 0
    missing_dirs = 0
    skipped_existing_links = 0

    for category_name, relative_dir in CATEGORY_MAPPINGS:
        print(f"\nCategory: {category_name}")

        category_id = category_ids.get(category_name)

        if category_id is None:
            created = create_category(server_url, category_name, token)
            if created == "CONFLICT":
                category_ids = fetch_categories(server_url, token)
                category_id = category_ids.get(category_name)
            elif created is None:
                category_id = None
            else:
                category_id = created
                category_ids[category_name] = category_id
                created_categories += 1
                print(f"  Created category with ID: {category_id}")
        else:
            print(f"  Category already exists with ID: {category_id}")

        if category_id is None:
            print("  Failed to create/find category. Skipping uploads for this category.")
            continue

        source_category_name = SHARED_CATEGORY_SOURCES.get(category_name)
        if source_category_name is not None:
            source_category_id = category_ids.get(source_category_name)
            if source_category_id is None:
                print(
                    f"  Missing source category '{source_category_name}'. Skipping shared linking.")
                continue

            source_sprites = fetch_cosmetics_by_category(
                server_url, source_category_id, token)
            if not source_sprites:
                print(
                    f"  No source sprites found in '{source_category_name}'. Skipping shared linking.")
                continue

            target_sprites = fetch_cosmetics_by_category(server_url, category_id, token)
            target_urls = {
                sprite.get("cosmetic_url")
                for sprite in target_sprites
                if sprite.get("cosmetic_url")
            }

            for source_sprite in source_sprites:
                source_sprite_id = source_sprite.get("cosmetic_id")
                source_sprite_url = source_sprite.get("cosmetic_url")

                if source_sprite_id is None:
                    continue

                if source_sprite_url in target_urls:
                    skipped_existing_links += 1
                    continue

                success, reason = link_sprite_by_id(
                    server_url, category_id, source_sprite_id, token)
                if success:
                    linked_sprites += 1
                    print(f"  Linked sprite: {source_sprite_id}")
                else:
                    failed_uploads += 1
                    print(f"  Failed linking sprite {source_sprite_id} ({reason})")

            continue

        sprite_dir = sprites_base_path / relative_dir
        if not sprite_dir.exists() or not sprite_dir.is_dir():
            missing_dirs += 1
            print(f"  Directory not found: {sprite_dir}")
            continue

        png_files = sorted(sprite_dir.glob("*.png"))
        if not png_files:
            print(f"  No PNG files found in: {sprite_dir}")
            continue

        for sprite_path in png_files:
            success, reason = upload_sprite(
                server_url, category_id, sprite_path, token)
            if success:
                uploaded_sprites += 1
                print(f"  Uploaded: {sprite_path.name}")
            else:
                failed_uploads += 1
                print(f"  Failed: {sprite_path.name} ({reason})")

    print("\nDone.")
    print(f"Categories created: {created_categories}")
    print(f"Sprites uploaded: {uploaded_sprites}")
    print(f"Sprites linked by ID: {linked_sprites}")
    print(f"Sprite upload failures: {failed_uploads}")
    print(f"Missing category directories: {missing_dirs}")
    print(f"Skipped existing links: {skipped_existing_links}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
