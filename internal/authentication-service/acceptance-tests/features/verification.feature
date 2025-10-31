Feature: Account Email Verification
  As a player
  I want to verify my account
  So that I can play the game

  Background:
    Given a player registers an account with the email "player@example.com" and password "Passw0rd1"
    And the password meets the required policy of at least 8 characters, including one number and one letter

  @wip
  Scenario: AC-1 Email verification code is sent after registration
    When the registration is completed
    Then an email should be sent to "player@example.com" containing a one-time verification code

  @wip
  Scenario: AC-2 Verify account successfully
    Given the player has received a verification email with a valid one-time code
    When the player submits the correct code within the valid time window
    Then the account should be marked as verified
    And the player should be able to log in successfully

  @wip
  Scenario: AC-3 Unable to login without verification
    Given the player has registered but not yet verified their account
    When the player attempts to log in to the game
    Then the login should be rejected
    And the player should see the message "Please verify your account before logging in"

  @wip
  Scenario: AC-4 Unable to verify account with invalid or expired code
    Given the player has received a verification email
    When the player enters an incorrect code
    Then the verification should fail
    And the player should see the message "Invalid verification code"

  @wip
  Scenario: AC-4b Unable to verify account after code expiration
    Given the player has received a verification email
    When the player enters the correct code after the time window has expired
    Then the verification should fail
    And the player should see the message "Verification code has expired"
