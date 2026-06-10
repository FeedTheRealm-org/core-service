import json
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
    return resp.json().get("data", {}).get("access_token")


@given("another user logs in")
def step_other_user_logs_in(context):
    email = f"zones_other_{uuid.uuid4().hex[:8]}@example.com"
    context.other_token = _signup_verify_login(email, "Password123")


@when('I publish zone "{zone_id}" with data:')
def step_publish_zone(context, zone_id):
    payload = {"data": json.loads(context.text.strip()) if context.text else {}}
    context.response = requests.put(
        f"{BASE_URL}/world/{context.world_id}/zones/{zone_id}",
        json=payload,
        headers=_auth_headers(context.token),
    )


@when('that user tries to publish zone "{zone_id}"')
def step_other_user_publish_zone(context, zone_id):
    payload = {"data": {"zone": "unauthorized"}}
    context.response = requests.put(
        f"{BASE_URL}/world/{context.world_id}/zones/{zone_id}",
        json=payload,
        headers=_auth_headers(context.other_token),
    )


@then('the world zones list should include zones "{zone_id_a}" and "{zone_id_b}"')
def step_world_zones_list_includes(context, zone_id_a, zone_id_b):
    resp = requests.get(
        f"{BASE_URL}/world/{context.world_id}/zones",
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to list zones: {resp.text}"
    zones = resp.json().get("data", {}).get("zones", [])
    zone_ids = {z.get("zone_id") for z in zones}
    assert int(zone_id_a) in zone_ids, f"Zone {zone_id_a} not in {zone_ids}"
    assert int(zone_id_b) in zone_ids, f"Zone {zone_id_b} not in {zone_ids}"


@then('the zone "{zone_id}" data should include "{value}"')
def step_zone_data_includes(context, zone_id, value):
    resp = requests.get(
        f"{BASE_URL}/world/{context.world_id}/zones/{zone_id}",
        headers=_auth_headers(context.token),
    )
    assert resp.status_code == 200, f"Failed to fetch zone: {resp.text}"
    zone_data = resp.json().get("data", {}).get("zone_data", "")
    assert value in zone_data, f"Expected '{value}' in zone_data: {zone_data}"
