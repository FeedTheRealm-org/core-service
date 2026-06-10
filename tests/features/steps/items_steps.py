import os
import uuid
import requests
from behave import when, then

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")


def _auth_headers(token):
    return {"Authorization": f"Bearer {token}"}


@when("I upload an item sprite")
def step_upload_item(context):
    item_id = str(uuid.uuid4())
    files = {
        "sprites[0]": ("sprite.png", b"dummy sprite", "image/png"),
    }
    data = {
        "ids[0]": item_id,
    }
    context.response = requests.put(
        f"{BASE_URL}/assets/items/world/{context.world_id}",
        headers=_auth_headers(context.token),
        files=files,
        data=data,
    )
    assert context.response.status_code == 201, (
        f"Upload item failed: {context.response.status_code} {context.response.text}"
    )
    context.item_id = item_id


@then("the item should be listed for the world")
def step_item_listed_for_world(context):
    resp = requests.get(
        f"{BASE_URL}/assets/items/world/{context.world_id}",
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to list items: {resp.text}"
    items = resp.json().get("data", {}).get("items", [])
    assert any(item.get("id") == context.item_id for item in items), (
        f"Item {context.item_id} not found in list: {items}"
    )


@then("I can fetch the item by id")
def step_fetch_item_by_id(context):
    resp = requests.get(
        f"{BASE_URL}/assets/items/{context.item_id}",
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to fetch item: {resp.text}"
    data = resp.json().get("data", {})
    assert data.get("id") == context.item_id, (
        f"Expected item {context.item_id}, got {data.get('id')}"
    )


@when("I delete the item")
def step_delete_item(context):
    context.response = requests.delete(
        f"{BASE_URL}/assets/items/{context.item_id}",
        headers=_auth_headers(context.token),
    )


@then("the item delete response should be successful")
def step_item_delete_success(context):
    assert context.response.status_code == 200, (
        f"Delete item failed: {context.response.status_code} {context.response.text}"
    )
    data = context.response.json().get("data", {})
    assert data.get("id") == context.item_id, (
        f"Expected deleted item id {context.item_id}, got {data.get('id')}"
    )
