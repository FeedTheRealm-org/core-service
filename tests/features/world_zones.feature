Feature: World Zones Management
  As a creator
  I want to publish and retrieve world zones
  So that servers can load the correct data

  Background:
    Given I have logged in with email "zones_owner@example.com" and password "Password123"

  Scenario: AC-1 Publish and retrieve zones
    Given I published a world with the name "zone.world"
    When I publish zone "1" with data:
      """
      {"zone":"alpha"}
      """
    And I publish zone "2" with data:
      """
      {"zone":"beta"}
      """
    Then the world zones list should include zones "1" and "2"
    And the zone "2" data should include "beta"

  Scenario: AC-2 Update existing zone data
    Given I published a world with the name "zone.update"
    When I publish zone "1" with data:
      """
      {"zone":"v1"}
      """
    And I publish zone "1" with data:
      """
      {"zone":"v2"}
      """
    Then the zone "1" data should include "v2"

  Scenario: AC-3 Non-owner cannot publish zone
    Given I published a world with the name "zone.private"
    And another user logs in
    When that user tries to publish zone "1"
    Then the response status should be 401
