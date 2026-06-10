Feature: Cosmetics Readonly
  As a player
  I want to view cosmetics catalog information
  So that I can browse available cosmetics

  Background:
    Given I have logged in with email "cosmetics_player@example.com" and password "Password123"

  Scenario: AC-1 List cosmetic categories
    When I request cosmetics categories
    Then the categories response should be a list

  Scenario: AC-2 Get cosmetics economy summary
    When I request the cosmetics economy summary
    Then the economy summary should include counts

  Scenario: AC-3 List cosmetics for a world
    Given I published a world with the name "cosmetics.world"
    When I request cosmetics for the world
    Then the cosmetics list response should be valid
