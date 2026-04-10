from behave import given, when, then
import requests

BASE_URL = "http://app:8000"

def _signup_verify_login(email, password):
    requests.post(f"{BASE_URL}/auth/signup", json={"email": email, "password": password})
    requests.post(f"{BASE_URL}/auth/verify", json={"email": email, "code": "IIIIIIII"})
    resp = requests.post(f"{BASE_URL}/auth/login", json={"email": email, "password": password})
    if resp.status_code == 200:
        data = resp.json().get("data", {})
        return data.get("access_token"), data.get("id")
    return None, None

@given('I have logged in with email "{email}" and password "{password}"')
def step_logged_in(context, email, password):
    token, player_id = _signup_verify_login(email, password)
    context.token = token
    context.player_id = player_id
    context.char_info = {
        "character_name": "DefaultName",
        "character_bio": "DefaultBio",
        "category_sprites": {
            "31174086-cd99-44db-9012-fbd2821f24c0": "31174086-cd99-44db-9012-fbd2821f24c0"
        }
    }

@when('I change my character name to "{name}"')
@when('I change my character name to "{name}" # less than {min} or more than {max} chars')
def step_change_name(context, name, min=None, max=None):
    context.char_info["character_name"] = name
    context.response = requests.patch(
        f"{BASE_URL}/player/character",
        json=context.char_info,
        headers={"Authorization": f"Bearer {context.token}"}
    )

@then('my character name should be updated')
def step_name_updated(context):
    assert context.response.status_code == 200, f"Failed: {context.response.text}"
    resp = requests.get(
        f"{BASE_URL}/player/character",
        headers={"Authorization": f"Bearer {context.token}"}
    )
    assert resp.status_code == 200
    data = resp.json().get("data", {})
    assert data.get("character_name") == context.char_info["character_name"]

@then('other players should see the updated name')
def step_other_players_name(context):
    other_token, _ = _signup_verify_login(f"other_{context.player_id}@example.com", "Password123")
    resp = requests.get(
        f"{BASE_URL}/player/character/{context.player_id}",
        headers={"Authorization": f"Bearer {other_token}"}
    )
    data = resp.json().get("data", {})
    assert data.get("character_name") == context.char_info["character_name"]

@then('I should see an error message "{msg}"')
def step_see_error(context, msg):
    body = context.response.json()
    detail = body.get("detail", "")
    assert msg in detail, f"Expected '{msg}' in '{detail}', body: {body}"

@when('I update my character bio to "{bio}"')
def step_update_bio(context, bio):
    context.char_info["character_bio"] = bio
    context.response = requests.patch(
        f"{BASE_URL}/player/character",
        json=context.char_info,
        headers={"Authorization": f"Bearer {context.token}"}
    )

@then('my character bio should be updated')
def step_bio_updated(context):
    assert context.response.status_code == 200, f"Failed: {context.response.text}"
    resp = requests.get(
        f"{BASE_URL}/player/character",
        headers={"Authorization": f"Bearer {context.token}"}
    )
    assert resp.status_code == 200
    data = resp.json().get("data", {})
    assert data.get("character_bio") == context.char_info["character_bio"]

@then('the updated bio should be visible to other players later')
def step_other_players_bio(context):
    other_token, _ = _signup_verify_login(f"other_bio_{context.player_id}@example.com", "Password123")
    resp = requests.get(
        f"{BASE_URL}/player/character/{context.player_id}",
        headers={"Authorization": f"Bearer {other_token}"}
    )
    data = resp.json().get("data", {})
    assert data.get("character_bio") == context.char_info["character_bio"]

@when('I update my character bio to a text longer than {limit:d} characters')
def step_bio_longer_than(context, limit):
    long_bio = "A" * (limit + 1)
    context.char_info["character_bio"] = long_bio
    context.response = requests.patch(
        f"{BASE_URL}/player/character",
        json=context.char_info,
        headers={"Authorization": f"Bearer {context.token}"}
    )
