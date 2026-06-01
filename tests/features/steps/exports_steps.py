import io
import os
import zipfile
import requests
from behave import given, when, then

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")
ADMIN_EMAIL = os.getenv("ADMIN_EMAIL", "")
ADMIN_PASSWORD = os.getenv("ADMIN_PASSWORD", "")


def _auth_headers(token):
    return {"Authorization": f"Bearer {token}"}


def _admin_login(context):
    assert ADMIN_EMAIL and ADMIN_PASSWORD, "ADMIN_EMAIL/ADMIN_PASSWORD must be set"
    resp = requests.post(
        f"{BASE_URL}/auth/login",
        json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD},
    )
    assert resp.status_code == 200, f"Admin login failed: {resp.text}"
    data = resp.json().get("data", {})
    token = data.get("access_token")
    assert token, f"Missing access_token in admin login response: {data}"
    context.admin_token = token
    context.token = token


def _build_zip_bytes():
    buffer = io.BytesIO()
    with zipfile.ZipFile(buffer, "w", zipfile.ZIP_DEFLATED) as archive:
        archive.writestr("readme.txt", "test export")
    buffer.seek(0)
    return buffer


@given("I have logged in as admin")
def step_login_admin(context):
    _admin_login(context)


@when('I query exports with app "{app}" os "{os_name}" version "{version}"')
def step_query_exports(context, app, os_name, version):
    context.response = requests.get(
        f"{BASE_URL}/exports/zip",
        params={"app": app, "os": os_name, "version": version},
    )


@when('I upload an export zip for app "{app}" version "{version}" os "{os_name}"')
def step_upload_export_zip(context, app, version, os_name):
    if not hasattr(context, "admin_token"):
        _admin_login(context)

    # Best-effort cleanup for idempotency.
    delete_resp = requests.delete(
        f"{BASE_URL}/exports/zip",
        params={"app": app, "os": os_name, "version": version},
        headers=_auth_headers(context.admin_token),
    )
    if delete_resp.status_code not in (204, 404):
        raise AssertionError(
            f"Unexpected delete status {delete_resp.status_code}: {delete_resp.text}"
        )

    zip_bytes = _build_zip_bytes()
    files = {
        "file": ("export.zip", zip_bytes, "application/zip"),
    }
    data = {
        "app": app,
        "version": version,
        "os": os_name,
        "release_note": "test export",
    }
    context.response = requests.put(
        f"{BASE_URL}/exports/zip",
        headers=_auth_headers(context.admin_token),
        files=files,
        data=data,
    )


@then('the export zip response should include version "{version}"')
def step_export_zip_response_version(context, version):
    resp = context.response
    assert resp.status_code == 201, (
        f"Expected 201, got {resp.status_code}: {resp.text}"
    )
    data = resp.json().get("data", {})
    assert data.get("version") == version, (
        f"Expected version '{version}', got '{data.get('version')}'"
    )
