#!/usr/bin/env python3
import sys
import time
import random
import requests
import os

BASE_URL = "http://localhost:8000"
WORDS_FILE = os.path.join(os.path.dirname(__file__), "world_name_words.csv")

# Cache for word lists
_word_cache = None


def load_word_lists():
    """Load word lists from file and cache them"""
    global _word_cache
    if _word_cache is not None:
        return _word_cache

    adjectives = []
    nouns = []
    suffixes = []

    try:
        print(f"Reading word lists from {WORDS_FILE}...", file=sys.stderr)
        with open(WORDS_FILE, "r", encoding="utf-8") as f:
            lines = f.readlines()

        # Skip header and comments
        data_lines = []
        for line in lines:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            # Skip the header row
            if line == "type,word":
                continue
            data_lines.append(line)

        for line in data_lines:
            parts = line.split(",", 1)  # Split on first comma only
            if len(parts) == 2:
                word_type, word = parts
                if word_type == "adjective":
                    adjectives.append(word)
                elif word_type == "noun":
                    nouns.append(word)
                elif word_type == "suffix":
                    suffixes.append(word)
        print(
            f"Loaded {len(adjectives)} adjectives, {len(nouns)} nouns, {len(suffixes)} suffixes from CSV",
            file=sys.stderr,
        )
    except FileNotFoundError:
        print(
            f"Warning: {WORDS_FILE} not found, using minimal fallback words",
            file=sys.stderr,
        )
        adjectives = ["Ancient", "Dark", "Mystical"]
        nouns = ["Realm", "World", "Land"]
        suffixes = ["ia", "land", "burg"]
    except Exception as e:
        print(f"Error reading words file: {e}, using fallback", file=sys.stderr)
        adjectives = ["Ancient", "Dark", "Mystical"]
        nouns = ["Realm", "World", "Land"]
        suffixes = ["ia", "land", "burg"]

    _word_cache = {"adjectives": adjectives, "nouns": nouns, "suffixes": suffixes}

    return _word_cache


def generate_random_name():
    """Generate a fantasy-like world name using word lists from file"""
    words = load_word_lists()

    adjectives = words["adjectives"]
    nouns = words["nouns"]
    suffixes = words["suffixes"]

    if not adjectives or not nouns:
        return f"World{random.randint(100, 999)}"  # Fallback

    adjective = random.choice(adjectives)
    noun = random.choice(nouns)

    if random.random() < 0.4 and suffixes:  # 40% chance for compound names
        name = f"{adjective} {noun}"
    else:
        suffix = random.choice(suffixes) if suffixes else ""
        name = f"{adjective}{noun}{suffix}"

    # Ensure reasonable length
    if len(name) > 24:
        name = name[:24].rstrip()

    return name


def get_token(email, password):
    try:
        r = requests.post(
            f"{BASE_URL}/auth/login",
            json={"email": email, "password": password},
            timeout=10,
        )
        r.raise_for_status()
        auth_json = r.json()
        token = auth_json["data"]["access_token"]
        if not token:
            print(f"Auth response did not contain token: {auth_json}", file=sys.stderr)
            sys.exit(1)
        return token
    except Exception as e:
        print(f"Authentication failed: {e}", file=sys.stderr)
        sys.exit(1)


def get_world_names(count):
    """Generate the specified number of world names"""
    names = []
    for _ in range(count):
        name = generate_random_name()
        names.append(name)
    return names


def post_worlds(token, count):
    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}",
    }

    # Get the required number of world names
    world_names = get_world_names(count)

    for i in range(0, count):
        world_name = world_names[i]
        # Create a safe filename by replacing spaces and special chars
        safe_filename = "".join(
            c for c in world_name if c.isalnum() or c in (" ", "-")
        ).rstrip()
        safe_filename = safe_filename.replace(" ", "_").lower()
        if not safe_filename:
            safe_filename = f"world_{i}"

        payload = {
            "data": {
                "worldName": world_name,
                "objectPlacementData": [
                    {"Position": {"x": -4, "y": 0, "z": -4}, "AssetDataId": 12},
                    {"Position": {"x": 0, "y": 0, "z": -5}, "AssetDataId": 4},
                ],
            },
            "file_name": f"{safe_filename}.world",
        }

        try:
            resp = requests.post(
                f"{BASE_URL}/world", json=payload, headers=headers, timeout=10
            )
        except Exception as e:
            print(f"[{i}] request failed: {e}", file=sys.stderr)
            continue

        status = resp.status_code
        body = None
        try:
            body = resp.json()
        except Exception:
            body = resp.text

        print(f"[{i}] {status} -> {body}")
        time.sleep(0.2)


if __name__ == "__main__":
    if len(sys.argv) < 4:
        print(
            "Usage: python3 ./generate_worlds.py <email> <password> <number_of_worlds>",
            file=sys.stderr,
        )
        sys.exit(1)

    email = sys.argv[1]
    password = sys.argv[2]
    try:
        count = int(sys.argv[3])
    except ValueError:
        print("number_of_worlds must be an integer", file=sys.stderr)
        sys.exit(1)

    token = get_token(email, password)
    post_worlds(token, count)
