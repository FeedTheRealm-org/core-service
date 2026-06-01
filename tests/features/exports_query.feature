Feature: Export Zip Queries
  As a client
  I want to query export zips
  So that I can download the correct build

  Scenario: AC-1 Invalid app name returns error
    When I query exports with app "invalid_app" os "linux" version "v1.0.0"
    Then the response status should be 400
    And the response should include an error message "app must be one of: ftr_world_editor, ftr_game"

  Scenario: AC-2 Invalid OS returns error
    When I query exports with app "ftr_game" os "macos" version "v1.0.0"
    Then the response status should be 400
    And the response should include an error message "os must be one of: linux, windows"

  Scenario: AC-3 Missing export returns not found
    When I query exports with app "ftr_game" os "linux" version "v0.0.1"
    Then the response status should be 404
    And the response should include an error message "export zip not found"
