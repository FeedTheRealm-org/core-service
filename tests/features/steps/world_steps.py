from behave import given, then
import requests

BASE_URL = "http://app:8000"


@given('I publish a world with name "{name}"')
@given('I publish a world with name ""')
def step_impl(context, name=""):
    payload = {"file_name": name, "data": {"worldName": name}}
    context.response = requests.post(
        f"{BASE_URL}/world",
        json=payload,
        headers={"Authorization": f"Bearer {context.token}"},
    )


@then("the world should be published")
def step_impl(context):
    assert context.response.status_code == 201, (
        f"Expected 201, got {context.response.status_code}: {context.response.text}"
    )
    context.world_id = context.response.json().get("data", {}).get("id")


@then("other players should see the world in the world listings")
def step_impl(context):
    # Register/login another player
    other_email = f"other_player_{context.player_id}@example.com"
    requests.post(
        f"{BASE_URL}/auth/signup",
        json={"email": other_email, "password": "Password123"},
    )
    requests.post(
        f"{BASE_URL}/auth/verify", json={"email": other_email, "code": "IIIIIIII"}
    )
    resp_login = requests.post(
        f"{BASE_URL}/auth/login", json={"email": other_email, "password": "Password123"}
    )
    assert resp_login.status_code == 200, f"Failed to login: {resp_login.text}"
    other_token = resp_login.json().get("data", {}).get("access_token")

    # Fetch worlds
    resp = requests.get(
        f"{BASE_URL}/world?offset=0&limit=10",
        headers={"Authorization": f"Bearer {other_token}"},
    )
    assert resp.status_code == 200, f"Expected 200, got {resp.status_code}"

    worlds = resp.json().get("data", {}).get("worlds", [])
    assert any(w.get("id") == context.world_id for w in worlds), (
        f"World not found in list: {worlds}"
    )
