Feature: World Access Tokens
  As a player
  I want to obtain a world join token
  So that I can connect to a world server

  Background:
    Given I have logged in with email "access_player@example.com" and password "Password123"

  Scenario: AC-1 Issue and consume a token
    Given I published a world with the name "access.world"
    And I have a character profile
    When I request a world join token
    Then I should receive a token
    When I consume the world join token
    Then the token should map to my user and world

  Scenario: AC-2 Cannot issue a token without a character
    Given I published a world with the name "access.nocharacter"
    When I request a world join token without a character
    Then the response status should be 403

  Scenario: AC-3 Invalid token cannot be consumed
    Given I have a character profile
    When I consume an invalid world join token
    Then the response status should be 400
