import os
import requests
from behave import when

BASE_URL = os.getenv("APP_BASE_URL", "http://app:8000")


@when('I query exports with app "{app}" os "{os_name}" version "{version}"')
def step_query_exports(context, app, os_name, version):
    context.response = requests.get(
        f"{BASE_URL}/exports/zip",
        params={"app": app, "os": os_name, "version": version},
    )
