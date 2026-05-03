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
DEFAULT_MODELS_DIR = "DefaultModels"


def usage() -> None:
    print("Usage: seed_default_models.py <SERVER_URL> <BASE_PATH>")
    print(f"  Looks for .glb files in <BASE_PATH>/{DEFAULT_MODELS_DIR}/")
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


def upload_model(server_url: str, model_path: Path, token: str):
    url = f"{server_url}/assets/models/world/{DEFAULT_WORLD_ID}"
    model_id = str(uuid.uuid4())

    with model_path.open("rb") as model_file:
        files = {"model_file": (model_path.name, model_file, "model/gltf-binary")}
        data = {"model_id": model_id}
        try:
            response = requests.put(
                url,
                files=files,
                data=data,
                headers=build_headers(token),
                timeout=60,
            )
        except requests.RequestException as exc:
            return False, str(exc), model_id

    if response.status_code in (200, 201):
        return True, "", model_id
    return False, f"HTTP {response.status_code} - {response.text}", model_id


def main() -> int:
    if len(sys.argv) != 3:
        usage()
        return 1

    server_url = sys.argv[1].rstrip("/")
    base_path = Path(sys.argv[2])

    if not base_path.exists() or not base_path.is_dir():
        print(f"Error: base path not found: {base_path}")
        return 1

    models_dir = base_path / DEFAULT_MODELS_DIR
    if not models_dir.exists() or not models_dir.is_dir():
        print(f"Error: '{DEFAULT_MODELS_DIR}' directory not found in: {base_path}")
        return 1

    token = get_token()

    glb_files = sorted(models_dir.glob("*.glb"))
    if not glb_files:
        print(f"No .glb files found in: {models_dir}")
        return 1

    print(
        f"Found {len(glb_files)} .glb files in '{models_dir}'. Uploading to world {DEFAULT_WORLD_ID}...\n"
    )

    uploaded = 0
    failed = 0

    for model_path in glb_files:
        success, reason, model_id = upload_model(server_url, model_path, token)
        if success:
            uploaded += 1
            print(f"  ✓ Uploaded: {model_path.name} (id: {model_id})")
        else:
            failed += 1
            print(f"  ✗ Failed:   {model_path.name} ({reason})")

    print(f"\nDone.")
    print(f"Uploaded: {uploaded}")
    print(f"Failed:   {failed}")

    return 0 if failed == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
