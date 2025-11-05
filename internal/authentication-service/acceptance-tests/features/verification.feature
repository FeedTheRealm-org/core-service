Feature: Account Email Verification
  As a player
  I want to verify my account
  So that I can play the game

  Scenario: AC-1 Email verification code is sent after registration
    Given a player registers an account with the email "player@example.com" and password "Passw0rd1"
    When the registration is completed
    Then an email should be sent to "player@example.com" containing a one-time verification code

  Scenario: AC-2 Verify account successfully
    Given a player registers an account with the email "player2@example.com" and password "Passw0rd1"
    When the player submits the correct code within the valid time window
    Then the account should be marked as verified
    And the player should be able to log in successfully

  Scenario: AC-3 Unable to login without verification
    Given a player registers an account with the email "not_verifid@example.com" and password "Passw0rd1"
    When the player attempts to log in to the game
    Then the login should be rejected
    And the player should see the message "Please verify your account before logging in"

  Scenario: AC-4 Unable to verify account with invalid or expired code
    Given a player registers an account with the email "invalid_code@example.com" and password "Passw0rd1"
    When the player enters an incorrect code
    Then the verification should fail
    And the player should see the message "Invalid verification code"
