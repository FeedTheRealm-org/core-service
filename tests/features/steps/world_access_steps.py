import os
import uuid
import requests
from behave import given, when, then

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")
STATIC_VERIFICATION_CODE = "IIIIIIII"


def _auth_headers(token):
    return {"Authorization": f"Bearer {token}"}


def _signup_verify_login(email, password):
    requests.post(
        f"{BASE_URL}/auth/signup",
        json={"email": email, "password": password},
    )
    requests.post(
        f"{BASE_URL}/auth/verify",
        json={"email": email, "code": STATIC_VERIFICATION_CODE},
    )
    resp = requests.post(
        f"{BASE_URL}/auth/login",
        json={"email": email, "password": password},
    )
    assert resp.status_code == 200, f"Login failed: {resp.text}"
    data = resp.json().get("data", {})
    return data.get("access_token"), data.get("id")


@given("I have a character profile")
def step_have_character_profile(context):
    char_info = getattr(context, "char_info", None)
    if char_info is None:
        char_info = {
            "character_name": "DefaultName",
            "character_bio": "DefaultBio",
            "category_sprites": {
                "31174086-cd99-44db-9012-fbd2821f24c0": "31174086-cd99-44db-9012-fbd2821f24c0"
            },
        }
    char_info = dict(char_info)
    char_info["character_name"] = f"Player{context.player_id[:8]}"
    char_info["character_bio"] = "Ready for worlds"
    context.char_info = char_info
    resp = requests.patch(
        f"{BASE_URL}/player/character",
        json=context.char_info,
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to create character: {resp.text}"


@when("I request a world join token")
def step_request_world_join_token(context):
    context.response = requests.post(
        f"{BASE_URL}/player/world-access/token",
        json={"world_id": context.world_id},
        headers=_auth_headers(context.token),
    )


@when("I request a world join token without a character")
def step_request_world_join_token_without_character(context):
    email = f"access_nochar_{uuid.uuid4().hex[:8]}@example.com"
    token, player_id = _signup_verify_login(email, "Password123")
    context.token = token
    context.player_id = player_id
    context.response = requests.post(
        f"{BASE_URL}/player/world-access/token",
        json={"world_id": context.world_id},
        headers=_auth_headers(context.token),
    )


@then("I should receive a token")
def step_receive_world_join_token(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data", {})
    token_id = data.get("token_id")
    assert token_id, f"Expected token_id in response: {data}"
    context.token_id = token_id


@when("I consume the world join token")
def step_consume_world_join_token(context):
    context.response = requests.post(
        f"{BASE_URL}/player/world-access/token/consume",
        json={"token_id": context.token_id},
        headers=_auth_headers(context.token),
    )


@when("I consume an invalid world join token")
def step_consume_invalid_world_join_token(context):
    context.response = requests.post(
        f"{BASE_URL}/player/world-access/token/consume",
        json={"token_id": "not-a-uuid"},
        headers=_auth_headers(context.token),
    )


@then("the token should map to my user and world")
def step_token_maps_to_user_and_world(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data", {})
    assert data.get("user_id") == context.player_id, (
        f"Expected user_id {context.player_id}, got {data.get('user_id')}"
    )
    assert data.get("world_id") == context.world_id, (
        f"Expected world_id {context.world_id}, got {data.get('world_id')}"
    )
