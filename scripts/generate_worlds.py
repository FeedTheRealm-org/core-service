#!/usr/bin/env python3
import sys
import time
import random
import string
import requests

BASE_URL = "http://localhost:8000"


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


def generate_random_name():
    """Generate a random world name with 6-24 characters (alphanumeric, no spaces)"""
    length = random.randint(6, 24)
    return "".join(random.choices(string.ascii_letters + string.digits, k=length))


def post_worlds(token, count):
    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}",
    }

    for i in range(0, count):
        world_name = generate_random_name()
        payload = {
            "data": {
                "worldName": world_name,
                "objectPlacementData": [
                    {"Position": {"x": -4, "y": 0, "z": -4}, "AssetDataId": 12},
                    {"Position": {"x": 0, "y": 0, "z": -5}, "AssetDataId": 4},
                ],
            },
            "file_name": f"{world_name}.world",
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
