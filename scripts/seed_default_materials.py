#!/usr/bin/env python3

# Get jwt token with:
# curl -X POST localhost:8000/auth/login -H "Content-Type: application/json" -d '{"email": "admin@admin.admin", "password": "admin123"}' | jq -r '.data.access_token'

import getpass
import os
import sys
import uuid
from pathlib import Path

import requests

DEFAULT_WORLD_ID = "00000000-0000-0000-0000-000000000000"
DEFAULT_MATERIALS_DIR = "Assets/DefaultMaterials"
GROUND_MATERIAL_TYPE = 0


def usage() -> None:
    print("Usage: seed_default_materials.py <SERVER_URL> <BASE_PATH>")
    print(f"  Looks for .png/.jpg/.jpeg files in <BASE_PATH>/{DEFAULT_MATERIALS_DIR}/")
    print("  Reads token from JWT_TOKEN env var; if empty/unset, asks via stdin.")


def get_token() -> str:
    token = os.getenv("JWT_TOKEN", "")
    if token == "":
        token = getpass.getpass("JWT token (press Enter for none): ")
    return token


def build_headers(token: str) -> dict:
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def upload_material(server_url: str, material_path: Path, token: str):
    url = f"{server_url}/assets/materials/world/{DEFAULT_WORLD_ID}"
    material_id = str(uuid.uuid4())
    material_name = material_path.stem.replace("_", " ")

    ext = material_path.suffix.lower()
    mime = "image/jpeg" if ext in (".jpg", ".jpeg") else "image/png"

    with material_path.open("rb") as f:
        files = {"materials[0]": (material_path.name, f, mime)}
        data = {
            "ids[0]": material_id,
            "names[0]": material_name,
            "material_types[0]": GROUND_MATERIAL_TYPE,
        }
        try:
            response = requests.put(
                url,
                files=files,
                data=data,
                headers=build_headers(token),
                timeout=60,
            )
        except requests.RequestException as exc:
            return False, str(exc), material_id

    if response.status_code in (200, 201):
        return True, "", material_id
    return False, f"HTTP {response.status_code} - {response.text}", material_id


def main() -> int:
    if len(sys.argv) != 3:
        usage()
        return 1

    server_url = sys.argv[1].rstrip("/")
    base_path = Path(sys.argv[2])

    if not base_path.exists() or not base_path.is_dir():
        print(f"Error: base path not found: {base_path}")
        return 1

    materials_dir = base_path / DEFAULT_MATERIALS_DIR
    if not materials_dir.exists() or not materials_dir.is_dir():
        print(f"Error: '{DEFAULT_MATERIALS_DIR}' directory not found in: {base_path}")
        return 1

    token = get_token()

    extensions = {".png", ".jpg", ".jpeg"}
    material_files = sorted(
        f for f in materials_dir.iterdir() if f.suffix.lower() in extensions
    )

    if not material_files:
        print(f"No image files found in: {materials_dir}")
        return 1

    print(
        f"Found {len(material_files)} materials in '{materials_dir}'. Uploading to world {DEFAULT_WORLD_ID}...\n"
    )

    uploaded = 0
    failed = 0

    for material_path in material_files:
        success, reason, material_id = upload_material(server_url, material_path, token)
        if success:
            uploaded += 1
            print(f"  ✓ Uploaded: {material_path.name} (id: {material_id})")
        else:
            failed += 1
            print(f"  ✗ Failed:   {material_path.name} ({reason})")

    print(f"\nDone.")
    print(f"Uploaded: {uploaded}")
    print(f"Failed:   {failed}")

    return 0 if failed == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
