#!/usr/bin/env python3

import os
import sys
import random
import requests
import string

def get_token() -> str:
    token = os.getenv("JWT_TOKEN", "")
    if not token:
        print("JWT_TOKEN environment variable not set.")
        sys.exit(1)
    return token

def build_headers(token: str) -> dict:
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    headers["Content-Type"] = "application/json"
    return headers

def generate_random_name():
    adjectives = ["Brave", "Dark", "Epic", "Fierce", "Giant", "Holy", "Iron", "Jade"]
    nouns = ["Knight", "Mage", "Orc", "Elf", "Dragon", "Sword", "Shield", "Hunter"]
    return f"{random.choice(adjectives)}{random.choice(nouns)}"

def create_bot_accounts(server_url: str, admin_token: str, num_bots: int = 5):
    for i in range(1, num_bots + 1):
        email = f"bot_{i}@feedtherealm.world"
        password = "qwerty123"
        print(f"Creating account for {email}...")

        # Create account
        post_url = f"{server_url}/auth/signup"
        try:
            signup_res = requests.post(post_url, json={"email": email, "password": password}, headers=build_headers(admin_token))
        except Exception as e:
            print(f"Failed to create bot {email}: {e}")
            continue

        print(f"Signup response: {signup_res.status_code}")

        # Login
        login_url = f"{server_url}/auth/login"
        try:
            # Requirements: "when trying to login as a bot account they should include in the request a valid Admin token"
            login_res = requests.post(login_url, json={"email": email, "password": password}, headers=build_headers(admin_token))
            if login_res.status_code != 200:
                print(f"Failed to login bot {email}: {login_res.text}")
                continue

            bot_token = login_res.json()["data"]["access_token"]
            print(f"Successfully logged in bot {email}")

            # Fetch cosmetic categories and random pages
            categories_url = f"{server_url}/assets/cosmetics/categories"
            cat_res = requests.get(categories_url, headers=build_headers(bot_token))
            if cat_res.status_code != 200:
                print("Failed to fetch cosmetic categories")
                continue

            categories = cat_res.json()["data"]["category_list"]
            sprites = {}

            for cat in categories:
                if random.random() < 0.3:
                    continue  # Skip some categories randomly

                cat_id = cat["category_id"]
                cat_name = cat["category_name"]

                # Fetch first page to get total count
                page_res = requests.get(f"{server_url}/assets/cosmetics/categories/{cat_id}?limit=24&offset=0", headers=build_headers(bot_token))
                if page_res.status_code != 200:
                    continue

                page_data = page_res.json().get("data", {})
                cosmetics = page_data.get("cosmetics_list", [])

                if cosmetics:
                    chosen = random.choice(cosmetics)
                    sprites[cat_id] = chosen["cosmetic_id"]

            if not sprites:
                print(f"No cosmetics found to assign for {email}")
                continue

            # Update Character
            char_patch_url = f"{server_url}/player/character"
            char_name = generate_random_name()
            patch_data = {
                "character_name": char_name,
                "category_sprites": sprites
            }
            char_res = requests.patch(char_patch_url, json=patch_data, headers=build_headers(bot_token))
            if char_res.status_code not in (200, 204): # Assuming patch is 200 or 204
                print(f"Failed to create/update character for {email}: {char_res.text}")
            else:
                print(f"Character '{char_name}' created/updated successfully for {email}")

        except Exception as e:
            print(f"Error during bot setup {email}: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: seed_bot_accounts.py <SERVER_URL>")
        sys.exit(1)

    url = sys.argv[1].rstrip("/")
    tkn = get_token()
    create_bot_accounts(url, tkn)
