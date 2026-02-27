#!/usr/bin/python3

import requests
import sys
import time

def fetch_categories(server_url, auth_token=None):
    url = f"{server_url}/assets/cosmetics/categories"
    headers = {'Authorization': f'Bearer {auth_token}'} if auth_token else {}

    try:
        response = requests.get(url, headers=headers)
        if response.status_code == 200:
            data = response.json()
            categories = data.get('data', {}).get('category_list', [])

            if categories:
                print("\n=== Current Categories ===")
                for idx, category in enumerate(categories, 1):
                    print(f"{idx}. {category['category_name']} (ID: {category['category_id']})")
                print(f"\nTotal: {len(categories)} categories\n")
            else:
                print("\nNo categories found.\n")

            return categories
        else:
            print(f"Failed to fetch categories: {response.status_code}")
            print(f"Response: {response.text}")
            return None
    except Exception as e:
        print(f"Error fetching categories: {str(e)}")
        return None

def add_category(server_url, category_name, auth_token=None):
    url = f"{server_url}/assets/cosmetics/categories"
    headers = {'Content-Type': 'application/json'}
    if auth_token:
        headers['Authorization'] = f'Bearer {auth_token}'

    payload = {'category_name': category_name}

    try:
        response = requests.post(url, json=payload, headers=headers)

        if response.status_code == 200 or response.status_code == 201:
            data = response.json()
            category_data = data.get('data', {})
            print(f"✓ Created: {category_data.get('category_name')} (ID: {category_data.get('category_id')})\n")
            time.sleep(0.5)
            return True
        elif response.status_code == 409:
            print(f"✗ Category '{category_name}' already exists\n")
            return False
        else:
            print(f"✗ Failed: {response.status_code} - {response.text}\n")
            return False
    except Exception as e:
        print(f"✗ Error: {str(e)}\n")
        return False

def interactive_menu(server_url, auth_token=None):
    print("Existing categories:")

    fetch_categories(server_url, auth_token)

    print("Add new categories (q to exit):")

    name = ""
    while name != "q":
        name = input("Enter category name: ").strip()
        if name and name != "q":
            add_category(server_url, name, auth_token)
        elif not name:
            print("Category name cannot be empty.\n")

def main():
    if len(sys.argv) < 2:
        print("Usage: python manage_categories.py <server_url> [auth_token]")
        print("Example (automatic): cat ./scripts/categories.txt | ./scripts/create_categories.py http://localhost:8000")
        print("Example (manual): python manage_categories.py http://localhost:8000")
        return

    server_url = sys.argv[1].rstrip('/')
    auth_token = sys.argv[2] if len(sys.argv) > 2 else None

    interactive_menu(server_url, auth_token)

if __name__ == '__main__':
    main()
