Feature: User Sign-Up
  As a new user
  I want to create an account
  So that I can authenticate and access the system

  @wip
  Scenario: AC-1 User can create an account with valid credentials
    When a sign-up request is made with email "alice@example.com" and password "SecurePass123!"
    Then the response should indicate success

  @wip
  Scenario: AC-2a Show error when email field is empty
    When a sign-up request is made with an empty email and password "Password123!"
    Then the response should include an error message "Email is required"

  @wip
  Scenario: AC-2b Show error when password field is empty
    When a sign-up request is made with email "charlie@example.com" and an empty password
    Then the response should include an error message "Password is required"

  @wip
  Scenario: AC-3 Prevent account creation with an existing email
    Given an account already exists with the email "existing@example.com"
    When a sign-up request is made with email "existing@example.com" and password "WonderPass123!"
    Then the response should include an error message "Email is already in use"
