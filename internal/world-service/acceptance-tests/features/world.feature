Feature: World Creation
  As a creator
  I want to publish my world
  So that other players can join it

  Background:
    Given I have logged in with email "test1@email.com" and password "Password123"

  Scenario: AC-1 Successfully publish world
    Given I publish a world with name "fantasy.realm"
    Then the world should be published
    And other players should see the world in the world listings

  Scenario: AC-2 Name length validation error
    Given I publish a world with name "Abc"
    Then I should see an error message "world name must be between 6 and 24 characters"
