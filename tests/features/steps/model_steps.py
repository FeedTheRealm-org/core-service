import uuid
import requests
from behave import given, when, then

BASE_URL = "http://app:8000"


@given('I published a world with the name "{name}"')
def step_impl(context, name):
    payload = {"file_name": name, "data": {"worldName": name}}
    resp = requests.post(
        f"{BASE_URL}/world",
        json=payload,
        headers={"Authorization": f"Bearer {context.token}"},
    )
    assert resp.status_code == 201, f"Failed to publish world: {resp.text}"
    context.world_id = resp.json().get("data", {}).get("id")
    assert context.world_id is not None


@when("I publish models related to the specified world")
def step_impl(context):
    world_id = context.world_id
    model_id = str(uuid.uuid4())
    context.last_model_id = model_id

    # We need multipart/form-data for models
    files = {
        "models[0].model_file": ("test.glb", b"dummy content", "model/gltf-binary")
    }
    data = {
        "models[0].model_id": model_id,
        "models[0].url": "http://example.com/test.glb",
    }

    context.response = requests.put(
        f"{BASE_URL}/assets/models/world/{world_id}",
        headers={"Authorization": f"Bearer {context.token}"},
        files=files,
        data=data,
    )


@then("the models should be published correctly")
def step_impl(context):
    assert context.response.status_code == 201, (
        f"Expected 201, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json()
    models_list = data.get("data", {}).get("models", [])
    assert len(models_list) > 0, f"Expected models list to not be empty, got {data}"


@when("I search for the world models by the world ID")
def step_search_world_models(context):
    world_id = context.world_id

    # We should have already published models (AC-2 requires it indirectly?)
    # AC-2 steps: "Given I published a world... When I search for the world models..."
    # Actually, if we haven't published models, the list might be empty. But AC-2 implies we published them.
    # Let me just publish a model right here first so the list won't be empty, if it wasn't published yet.
    if not hasattr(context, "last_model_id"):
        model_id = str(uuid.uuid4())
        context.last_model_id = model_id
        files = {
            "models[0].model_file": ("test.glb", b"dummy content", "model/gltf-binary")
        }
        data = {
            "models[0].model_id": model_id,
            "models[0].url": "http://example.com/test.glb",
        }
        requests.put(
            f"{BASE_URL}/assets/models/world/{world_id}",
            headers={"Authorization": f"Bearer {context.token}"},
            files=files,
            data=data,
        )

    context.response = requests.get(
        f"{BASE_URL}/assets/models/world/{world_id}",
        headers={"Authorization": f"Bearer {context.token}"},
    )


@given("I publish world models without a world ID")
def step_publish_without_world_id(context):
    context.response = requests.put(
        f"{BASE_URL}/assets/models/world/invalid_id",
        headers={"Authorization": f"Bearer {context.token}"},
    )


@then('I get the error "{error}"')
def step_get_error(context, error):
    body = context.response.json()
    detail = body.get("detail", "")
    assert error in detail, f"Expected '{error}' in '{detail}', body: {body}"


@then("I get the correct world models")
def step_correct_world_models(context):
    assert context.response.status_code == 200, (
        f"Expected 200, got {context.response.status_code}: {context.response.text}"
    )
    data = context.response.json()
    models_list = data.get("data", {}).get("models", [])
    assert len(models_list) > 0, f"Expected models list to not be empty, got {data}"


@when("I attempt to publish models without models")
def step_publish_without_models(context):
    world_id = context.world_id
    context.response = requests.put(
        f"{BASE_URL}/assets/models/world/{world_id}",
        headers={"Authorization": f"Bearer {context.token}"},
    )
