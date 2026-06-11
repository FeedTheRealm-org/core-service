import os
import requests
from behave import when, then

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")


def _auth_headers(token):
    return {"Authorization": f"Bearer {token}"}


@when("I request cosmetics categories")
def step_request_cosmetics_categories(context):
    context.response = requests.get(
        f"{BASE_URL}/assets/cosmetics/categories",
        headers=_auth_headers(context.token),
    )


@then("the categories response should be a list")
def step_categories_response_list(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data", {})
    categories = data.get("category_list")
    assert isinstance(categories, list), f"Expected list, got: {data}"


@when("I request the cosmetics economy summary")
def step_request_cosmetics_economy_summary(context):
    context.response = requests.get(
        f"{BASE_URL}/assets/cosmetics/economy-summary",
        headers=_auth_headers(context.token),
    )


@then("the economy summary should include counts")
def step_economy_summary_includes_counts(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data", {})
    assert "default_cosmetics" in data, f"Missing default_cosmetics: {data}"
    assert "user_created_cosmetics" in data, f"Missing user_created_cosmetics: {data}"
    assert "average_price" in data, f"Missing average_price: {data}"


@when("I request cosmetics for the world")
def step_request_cosmetics_for_world(context):
    context.response = requests.get(
        f"{BASE_URL}/assets/cosmetics/worlds/{context.world_id}",
        headers=_auth_headers(context.token),
        params={"offset": 0, "limit": 24},
    )


@then("the cosmetics list response should be valid")
def step_cosmetics_list_valid(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json().get("data", {})
    cosmetics_list = data.get("cosmetics_list")
    assert isinstance(cosmetics_list, list), f"Expected list, got: {data}"
