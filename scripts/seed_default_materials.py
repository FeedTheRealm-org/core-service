#!/usr/bin/env python3

# Get jwt token with:
# curl -X POST localhost:8000/auth/login -H "Content-Type: application/json" -d '{"email": "admin@admin.admin", "password": "admin123"}' | jq -r '.data.access_token'

import getpass
import os
import sys
import uuid
import re
from pathlib import Path

import requests

BATCH_SIZE = 10
GROUND_MATERIAL_TYPE = 0
DEFAULT_WORLD_ID = "00000000-0000-0000-0000-000000000000"
TEXTURES_DIR = "Assets/Cartoon_Texture_Pack"


def usage() -> None:
    print(f"Usage: {sys.argv[0]} <SERVER_URL> <BASE_PATH>")
    print("  Looks for .png files ending in Basecolor.png, Basecolor_A.png, etc.")
    print(f"  in <BASE_PATH>/{TEXTURES_DIR}/")
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


def format_material_name(filename: str, name_counters: dict) -> str:
    """
    Converts a filename like 'Grass_Dense_Tint_01_Base_Basecolor_A' to 'Grass 1',
    'Grass 2', etc. — incrementing per unique material type across all batches.
    """
    base_name = re.sub(r"_Basecolor(_[A-Z])?$", r"\1", filename, flags=re.IGNORECASE)
    base_name = re.sub(r"_?Basecolor$", "", base_name, flags=re.IGNORECASE)

    parts = base_name.split("_")
    material = parts[0].capitalize() if parts else "Unknown"

    name_counters[material] = name_counters.get(material, 0) + 1
    return f"{material} {name_counters[material]}"


def upload_materials(server_url: str, materials: list, token: str):
    url = f"{server_url}/assets/materials/world/{DEFAULT_WORLD_ID}"

    files_to_send = []
    data = {}

    print(f"Batch uploading {len(materials)} materials...")

    for idx, (path, name) in enumerate(materials):
        mat_id = str(uuid.uuid4())

        data[f"ids[{idx}]"] = mat_id
        data[f"material_types[{idx}]"] = GROUND_MATERIAL_TYPE
        data[f"names[{idx}]"] = name

        f = path.open("rb")
        files_to_send.append((f"materials[{idx}]", (path.name, f, "image/png")))
        print(f"  Preparing: {name} as {mat_id} ({path.name})")

    try:
        response = requests.put(
            url,
            files=files_to_send,
            data=data,
            headers=build_headers(token),
            timeout=120,
        )

        for _, (_, f, _) in files_to_send:
            f.close()

    except requests.RequestException as exc:
        for _, (_, f, _) in files_to_send:
            f.close()
        return False, str(exc), []

    if response.status_code in (200, 201):
        return True, "", [data[f"ids[{i}]"] for i in range(len(materials))]
    return False, f"HTTP {response.status_code} - {response.text}", []


def main() -> int:
    if len(sys.argv) != 3:
        usage()
        return 1

    server_url = sys.argv[1].rstrip("/")
    base_path = Path(sys.argv[2])

    if not base_path.exists() or not base_path.is_dir():
        print(f"Error: base path not found: {base_path}")
        return 1

    textures_dir = base_path / TEXTURES_DIR
    if not textures_dir.exists() or not textures_dir.is_dir():
        print(f"Error: '{TEXTURES_DIR}' directory not found in: {base_path}")
        return 1

    token = get_token()

    png_files = []
    for root, _, sorted_files in os.walk(textures_dir):
        for file in sorted_files:
            if file.endswith(".png") and ("Basecolor" in file or "basecolor" in file):
                png_files.append(Path(root) / file)

    png_files.sort()

    if not png_files:
        print(f"No Basecolor .png files found in: {textures_dir}")
        return 1

    print(
        f"Found {len(png_files)} Basecolor .png files. Uploading to world {DEFAULT_WORLD_ID}...\n"
    )

    uploaded_count = 0
    failed_count = 0

    name_counters: dict[str, int] = {}
    materials = []
    for path in png_files:
        name = format_material_name(path.stem, name_counters)
        materials.append((path, name))

    for i in range(0, len(materials), BATCH_SIZE):
        batch = materials[i : i + BATCH_SIZE]
        print(
            f"\n--- Batch {i // BATCH_SIZE + 1} of {(len(materials) - 1) // BATCH_SIZE + 1} ---"
        )

        success, reason, _ = upload_materials(server_url, batch, token)

        if success:
            uploaded_count += len(batch)
            print("  ✓ Batch successful")
        else:
            failed_count += len(batch)
            print(f"  ✗ Batch failed: {reason}")

    print("\nDone.")
    print(f"Uploaded: {uploaded_count}")
    print(f"Failed:   {failed_count}")

    return 0 if failed_count == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
