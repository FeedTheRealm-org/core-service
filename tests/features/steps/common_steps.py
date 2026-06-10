from behave import then


@then("the response status should be {status:d}")
def step_response_status(context, status):
    assert hasattr(context, "response"), "No response stored on context"
    assert context.response is not None, "No response stored on context"
    assert context.response.status_code == status, (
        f"Expected status {status}, got {context.response.status_code}: {context.response.text}"
    )
