Feature: User Sign-Up
  As a new user
  I want to create an account
  So that I can authenticate and access the system

  Scenario: AC-1 User can create an account with valid credentials
    When a sign-up request is made with email "alice@example.com" and password "SecurePass123!"
    Then the response should indicate success

  Scenario: AC-2a Show error when email field is empty
    When a sign-up request is made with an empty email and password "Password123!"
    Then the response should include an error message "Email is required"

  Scenario: AC-2b Show error when password field is empty
    When a sign-up request is made with email "charlie@example.com" and an empty password
    Then the response should include an error message "Password is required"

  Scenario: AC-3 Prevent account creation with an existing email
    Given an account already exists with the email "existing@example.com"
    When a sign-up request is made with email "existing@example.com" and password "WonderPass123!"
    Then the response should include an error message "Email is already in use"

  @wip
  Scenario: AC-4a Show error when the email format is invalid
    When a sign-up request is made with email "invalid-email" and password "StrongPass123!"
    Then the response should include an error message "Invalid email format"

  @wip
  Scenario: AC-4b Show error when the email domain is invalid
    When a sign-up request is made with email "user@invalid" and password "StrongPass123!"
    Then the response should include an error message "Invalid email format"

  @wip
  Scenario: AC-5a Show error when the password is too short
    When a sign-up request is made with email "shortpass@example.com" and password "Ab12"
    Then the response should include an error message "Password must be at least 8 characters long"

  @wip
  Scenario: AC-5b Show error when the password has no numbers
    When a sign-up request is made with email "nonumeric@example.com" and password "PasswordOnly"
    Then the response should include an error message "Password must contain at least one number"

  @wip
  Scenario: AC-5c Show error when the password has no letters
    When a sign-up request is made with email "noletters@example.com" and password "12345678"
    Then the response should include an error message "Password must contain at least one letter"
