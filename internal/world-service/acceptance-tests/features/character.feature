Feature: World Creation
  As a creator
  I want to publish my world
  So that other players can join it

  Background:
    Given I have logged in with email "test1@email.com" and password "Password123"

  Scenario: AC-1 Successfully publish world
    When I publish a world with name "Fantasy Realm"
    Then the world should be published
    And other players should see "Fantasy Realm" in the world listings

  Scenario: AC-2 Name length validation error
    When I publish a world with name "Abc" # less than 4 or more than 24 chars
    Then I should see an error message "Name must be between 4 and 24 characters"

  Scenario: AC-3 Successfully get published world details
    Given I have published a world with name "Fantasy Realm"
    When I request the world details for "Fantasy Realm"
    Then I should receive the world's details

  Scenario: AC-4 Successfully delete world
    Given I have published a world with name "Fantasy Realm"
    When I delete the world "Fantasy Realm"
    Then the world should be removed
    And other players should no longer see "Fantasy Realm" in the world listings
