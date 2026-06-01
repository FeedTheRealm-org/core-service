Feature: World Update
  As a creator
  I want to update world data and createable data
  So that players see the latest content

  Background:
    Given I have logged in with email "world_update@example.com" and password "Password123"

  Scenario: AC-1 Update world description and data
    Given I published a world with the name "update.world"
    When I update the world description to "new description" with data:
      """
      {"k":2}
      """
    Then the world details should reflect the updated description "new description"
    And the world data should include "\"k\":2"

  Scenario: AC-2 Update createable data
    Given I published a world with the name "createable.world"
    When I update the world createable data to:
      """
      {"c":5}
      """
    Then the world createable data should include "\"c\":5"
