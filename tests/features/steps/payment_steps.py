import os
import requests
from behave import when, then

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")


def _auth_headers(token):
    return {"Authorization": f"Bearer {token}"}


@when("I request my gem balance")
def step_request_gem_balance(context):
    context.response = requests.get(
        f"{BASE_URL}/payments/gems/balances",
        headers=_auth_headers(context.token),
    )


@then("I should receive my gem balance")
def step_receive_gem_balance(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data", {})
    assert data.get("user_id") == context.player_id, (
        f"Expected user_id {context.player_id}, got {data.get('user_id')}"
    )
    assert data.get("gems") is not None, f"Missing gems in response: {data}"
    assert int(data.get("gems")) >= 0, f"Expected non-negative gems: {data}"


@when("I request my creator balance")
def step_request_creator_balance(context):
    context.response = requests.get(
        f"{BASE_URL}/payments/balances/creators",
        headers=_auth_headers(context.token),
    )


@then("I should receive my creator balance")
def step_receive_creator_balance(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data", {})
    assert data.get("user_id") == context.player_id, (
        f"Expected user_id {context.player_id}, got {data.get('user_id')}"
    )
    assert data.get("balance") is not None, f"Missing balance in response: {data}"


@when("I request gem packs")
def step_request_gem_packs(context):
    context.response = requests.get(
        f"{BASE_URL}/payments/gems/packs",
        headers=_auth_headers(context.token),
    )


@then("I should receive the gem packs list")
def step_receive_gem_packs(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data")
    assert isinstance(data, list), f"Expected list of packs, got: {data}"
