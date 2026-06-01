Feature: World Items Upload
  As a creator
  I want to upload item sprites
  So that players can use them in worlds

  Background:
    Given I have logged in with email "items_owner@example.com" and password "Password123"

  Scenario: AC-1 Upload and retrieve items
    Given I published a world with the name "items.world"
    When I upload an item sprite
    Then the item should be listed for the world
    And I can fetch the item by id
    When I delete the item
    Then the item delete response should be successful
