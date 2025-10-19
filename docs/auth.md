# Authentication API Documentation

## Overview

The Authentication Service provides endpoints for user account management and authentication.

---

## Endpoints

### POST /auth/signup

Creates a new user account.

#### Request

**URL:** `/auth/signup`

**Method:** `POST`

**Content-Type:** `application/json`

**Body Parameters:**

| Parameter | Type   | Required | Description                    |
|-----------|--------|----------|--------------------------------|
| email     | string | Yes      | User's email address           |
| password  | string | Yes      | User's password                |

**Example Request:**

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
}
```

#### Response

**Success Response (201 Created):**

```json
{
  "email": "user@example.com",
  "message": "Account created successfully",
}
```

**Error Responses:**

| Status Code | Error Message              | Description                                    |
|-------------|----------------------------|------------------------------------------------|
| 400         | `Email is required`        | The email field is empty or missing            |
| 400         | `Password is required`     | The password field is empty or missing         |
| 400         | `Email is already in use`  | An account with this email already exists      |
| 500         | Internal server error      | An unexpected error occurred                   |

**Example Error Response:**

```json
{
  "error": "Email is already in use",
}
```

#### Usage Examples

```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```
