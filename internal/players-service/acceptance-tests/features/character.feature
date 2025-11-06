Feature: Player character info
  As a player
  I want to edit my character data
  So that other players can see it

  Background:
    Given I have logged in with email "test1@email.com" and password "Password123"

  Scenario: AC-1 Successfully update character name
    When I change my character name to "StormRider"
    Then my character name should be updated
    And other players should see the updated name

  @wip
  Scenario: AC-2 Name length validation error
    When I change my character name to "S" # less than 4 or more than 24 chars
    Then I should see an error message "Name must be between 4 and 24 characters"

  @wip
  Scenario: AC-3 Successfully update character bio
    When I update my character bio to "A brave explorer traveling the stars!"
    And I save the character profile
    Then my character bio should be updated
    And the updated bio should be visible to other players later

  @wip
  Scenario: AC-4 Bio length validation error
    When I update my character bio to a text longer than 150 characters
    And I save the character profile
    Then I should see an error message "Bio cannot exceed 150 characters"
