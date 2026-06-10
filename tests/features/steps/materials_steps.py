import os
import uuid
import requests
from behave import when, then

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")


def _auth_headers(token):
    return {"Authorization": f"Bearer {token}"}


@when('I upload a material named "{name}"')
def step_upload_material(context, name):
    material_id = str(uuid.uuid4())
    files = {
        "materials[0]": ("material.png", b"dummy material", "image/png"),
    }
    data = {
        "ids[0]": material_id,
        "names[0]": name,
    }
    context.response = requests.put(
        f"{BASE_URL}/assets/materials/world/{context.world_id}",
        headers=_auth_headers(context.token),
        files=files,
        data=data,
    )
    assert context.response.status_code == 201, (
        f"Upload material failed: {context.response.status_code} {context.response.text}"
    )
    context.material_id = material_id


@then("the material should appear in the materials list for the world")
def step_material_listed_for_world(context):
    resp = requests.get(
        f"{BASE_URL}/assets/materials",
        headers=_auth_headers(context.token),
        params={"world_id": context.world_id, "offset": 0, "limit": 200},
    )
    assert resp.status_code == 200, f"Failed to list materials: {resp.text}"
    data = resp.json().get("data", [])
    assert any(material.get("id") == context.material_id for material in data), (
        f"Material {context.material_id} not found in list: {data}"
    )


@when("I delete the material")
def step_delete_material(context):
    context.response = requests.delete(
        f"{BASE_URL}/assets/materials/{context.material_id}",
        headers=_auth_headers(context.token),
    )


@then("the material delete response should be successful")
def step_material_delete_success(context):
    assert context.response.status_code == 200, (
        f"Delete material failed: {context.response.status_code} {context.response.text}"
    )
    data = context.response.json().get("data", {})
    assert data.get("id") == context.material_id, (
        f"Expected deleted material id {context.material_id}, got {data.get('id')}"
    )
