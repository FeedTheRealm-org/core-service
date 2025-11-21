Feature: World Models Management
  As a creator
  I want to publish and retrieve models for a world
  So that other players can load them

  Background:
    Given I have logged in with email "test1@email.com" and password "Password123"

  Scenario: AC-1 Successfully publish world models
    Given I published a world with the name "fantasy.land"
    When I publish models related to the specified world
    Then the models should be published correctly

  Scenario: AC-2 Get World Models by world ID
    Given I published a world with the name "fantasy.land"
    When I search for the world models by the world ID
    Then I get the correct world models

  Scenario: AC-3 Cannot publish without World ID
    Given I publish world models without a world ID
    Then I get the error "world id is required"

  Scenario: AC-4 Cannot publish without models
    Given I published a world with the name "fantasy.land"
    When I attempt to publish models without models
    Then I get the error "models list cannot be empty"