#!/usr/bin/python3

import os
import requests
import sys
from pathlib import Path

def fetch_categories(server_url, auth_token=None):
    """Fetch available categories from the server."""
    url = f"{server_url}/assets/cosmetics/categories"
    headers = {'Authorization': f'Bearer {auth_token}'} if auth_token else {}

    try:
        response = requests.get(url, headers=headers)
        if response.status_code == 200:
            data = response.json()
            return data.get('data', {}).get('category_list', [])
        else:
            print(f"Failed to fetch categories: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Error fetching categories: {str(e)}")
        return None

def select_category(categories):
    """Let user select a category from the list."""
    print("\nAvailable categories:")
    for idx, category in enumerate(categories, 1):
        print(f"{idx}. {category['category_name']} (ID: {category['category_id']})")

    while True:
        try:
            choice = int(input("\nSelect category number: "))
            if 1 <= choice <= len(categories):
                selected = categories[choice - 1]
                print(f"Selected: {selected['category_name']}")
                return selected['category_id']
            else:
                print(f"Please enter a number between 1 and {len(categories)}")
        except ValueError:
            print("Please enter a valid number")

def get_file_extensions():
    """Ask user for file extensions to look for."""
    extensions_input = input("\nEnter file extensions to upload (comma-separated, e.g., png,jpg): ")
    extensions = [ext.strip().lstrip('.') for ext in extensions_input.split(',')]
    return extensions

def upload_sprite(server_url, file_path, category_id, auth_token=None):
    """Upload a single sprite file to the server."""
    base_url = f"{server_url}/assets/cosmetics/categories"

    ext = file_path.suffix.lower()
    mime_type = 'image/png' if ext == '.png' else 'image/jpeg'

    try:
        with open(file_path, 'rb') as f:
            files = {'sprite': (os.path.basename(file_path), f, mime_type)}
            data = {'category_id': category_id}
            headers = {'Authorization': f'Bearer {auth_token}'} if auth_token else {}

            response = requests.put(f"{base_url}/{category_id}", files=files, data=data, headers=headers)

            if response.status_code == 201:
                result = response.json()
                print(f"✓ Uploaded: {file_path.name} -> {result['data']['sprite_id']}")
                return True
            else:
                print(f"✗ Failed: {file_path.name} - {response.status_code} - {response.text}")
                return False
    except Exception as e:
        print(f"✗ Error uploading {file_path.name}: {str(e)}")
        return False

def main():
    if len(sys.argv) < 3:
        print("Usage: python upload_assets.py <server_url> <assets_folder_path> [auth_token]")
        print("Example: python upload_assets.py http://localhost:8000 ./cosmetics")
        return

    server_url = sys.argv[1].rstrip('/')
    assets_path = Path(sys.argv[2])
    auth_token = sys.argv[3] if len(sys.argv) > 3 else None

    if not assets_path.exists() or not assets_path.is_dir():
        print(f"Error: {assets_path} is not a valid directory")
        return

    categories = fetch_categories(server_url, auth_token)
    if not categories:
        print("No categories available or failed to fetch categories")
        return

    category_id = select_category(categories)

    extensions = get_file_extensions()
    all_files = []
    for ext in extensions:
        all_files.extend(assets_path.glob(f'*.{ext}'))

    if not all_files:
        print(f"No files found with extensions: {', '.join(extensions)}")
        return

    print(f"\nFound {len(all_files)} files to upload")
    print(f"Server: {server_url}")
    print(f"Category ID: {category_id}\n")

    confirm = input("Proceed with upload? (y/n): ")
    if confirm.lower() != 'y':
        print("Upload cancelled")
        return

    success_count = 0
    fail_count = 0

    for file_path in all_files:
        if upload_sprite(server_url, file_path, category_id, auth_token):
            success_count += 1
        else:
            fail_count += 1

    print(f"\n--- Summary ---")
    print(f"Successfully uploaded: {success_count}")
    print(f"Failed: {fail_count}")

if __name__ == '__main__':
    main()
