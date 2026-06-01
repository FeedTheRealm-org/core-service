import json
import os
import requests
from behave import when, then

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")


def _auth_headers(token):
    return {"Authorization": f"Bearer {token}"}


@when('I update the world description to "{description}" with data:')
def step_update_world_description(context, description):
    payload = {
        "file_name": "",
        "description": description,
        "data": json.loads(context.text.strip()) if context.text else {},
    }
    context.response = requests.put(
        f"{BASE_URL}/world/{context.world_id}",
        json=payload,
        headers=_auth_headers(context.token),
    )


@when("I update the world createable data to:")
def step_update_createable_data(context):
    payload = {
        "createable_data": json.loads(context.text.strip()) if context.text else {},
    }
    context.response = requests.put(
        f"{BASE_URL}/world/{context.world_id}/createable-data",
        json=payload,
        headers=_auth_headers(context.token),
    )


@then('the world details should reflect the updated description "{description}"')
def step_world_details_updated(context, description):
    assert context.response.status_code == 200, (
        f"Update world failed: {context.response.status_code} {context.response.text}"
    )
    resp = requests.get(
        f"{BASE_URL}/world/{context.world_id}",
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to get world: {resp.text}"
    data = resp.json().get("data", {})
    assert data.get("description") == description, (
        f"Expected description '{description}', got '{data.get('description')}'"
    )


@then('the world data should include "{fragment}"')
def step_world_data_includes(context, fragment):
    resp = requests.get(
        f"{BASE_URL}/world/{context.world_id}",
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to get world: {resp.text}"
    data_str = resp.json().get("data", {}).get("data", "")
    normalized_fragment = fragment.replace('\\"', '"')
    expected = None
    try:
        expected = json.loads("{" + normalized_fragment + "}")
    except json.JSONDecodeError:
        expected = None

    try:
        parsed = json.loads(data_str)
    except json.JSONDecodeError:
        parsed = None

    if expected is not None and isinstance(parsed, dict):
        for key, value in expected.items():
            assert parsed.get(key) == value, (
                f"Expected {key}={value} in world data: {data_str}"
            )
    else:
        normalized = "".join(str(data_str).split())
        target = "".join(str(normalized_fragment).split())
        assert target in normalized, f"Expected '{fragment}' in world data: {data_str}"


@then('the world createable data should include "{fragment}"')
def step_world_createable_data_includes(context, fragment):
    resp = requests.get(
        f"{BASE_URL}/world/{context.world_id}",
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to get world: {resp.text}"
    createable_str = resp.json().get("data", {}).get("createable_data", "")
    normalized_fragment = fragment.replace('\\"', '"')
    expected = None
    try:
        expected = json.loads("{" + normalized_fragment + "}")
    except json.JSONDecodeError:
        expected = None

    try:
        parsed = json.loads(createable_str)
    except json.JSONDecodeError:
        parsed = None

    if expected is not None and isinstance(parsed, dict):
        for key, value in expected.items():
            assert parsed.get(key) == value, (
                f"Expected {key}={value} in createable data: {createable_str}"
            )
    else:
        normalized = "".join(str(createable_str).split())
        target = "".join(str(normalized_fragment).split())
        assert target in normalized, (
            f"Expected '{fragment}' in createable data: {createable_str}"
        )
