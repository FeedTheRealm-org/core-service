Feature: Refresh Verification Code
  As a registered but unverified player
  I want to request a new verification code
  So that I can verify my account if the original code expired or was lost

  Scenario: AC-1 Successfully request a new verification code
    Given a player registers an account with the email "refresh1@example.com" and password "Passw0rd1"
    When the player requests a new verification code for "refresh1@example.com"
    Then the response should indicate the verification code was refreshed

  Scenario: AC-2 Cannot refresh verification code for non-existing email
    When the player requests a new verification code for "doesnotexist@example.com"
    Then the response should include a player error message "No account exists with the provided email address."

  Scenario: AC-3 Cannot refresh verification code for already verified account
    Given a player registers an account with the email "refresh3@example.com" and password "Passw0rd1"
    And the player verifies the account for "refresh3@example.com"
    When the player requests a new verification code for "refresh3@example.com"
    Then the response should include a player error message "The account is already verified; no verification code will be generated."

  Scenario: AC-4 Account can be verified after receiving a refreshed code
    Given a player registers an account with the email "refresh4@example.com" and password "Passw0rd1"
    And the player requests a new verification code for "refresh4@example.com"
    When the player submits the correct code for "refresh4@example.com"
    Then the account should be verified successfully
