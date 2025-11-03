Feature: User Login
  As a user with an existing account
  I want to complete the fields with my credentials
  So that I can access my account in the application

  Background:
    Given an account already exists with the email "login@example.com" and password "somepassword1"

  @wip
  Scenario: AC-1 Successful login with valid credentials
    When a login request is made with email "login@example.com" and password "somepassword1"
    Then the response should indicate a successful login

  Scenario: AC-2a Show error when password is incorrect
    When a login request is made with email "login@example.com" and password "WrongPass!"
    Then the response should include an error message "Invalid email or password"

  Scenario: AC-2b Show error when email does not exist
    When a login request is made with email "unknown@example.com" and password "SomePass123!"
    Then the response should include an error message "Invalid email or password"

  Scenario: AC-3a Show error when email field is empty
    When a login request is made with an empty email and password "ValidPass123!"
    Then the response should include an error message "Email is required"

  Scenario: AC-3b Show error when password field is empty
    When a login request is made with email "login@example.com" and an empty password
    Then the response should include an error message "Password is required"

  @wip
  Scenario: AC-4a Session remains active before timeout
    Given the user has logged in successfully
    When "30" minutes have passed since login
    Then the session should still be active

  Scenario: AC-4b Session expires after timeout
    Given the user has logged in successfully
    When "60" minutes have passed since login
    Then the session should be closed
    And further requests should require authentication
