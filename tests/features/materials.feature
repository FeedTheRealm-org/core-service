Feature: World Materials Upload
  As a creator
  I want to upload materials
  So that worlds can render correctly

  Background:
    Given I have logged in with email "materials_owner@example.com" and password "Password123"

  Scenario: AC-1 Upload and list materials
    Given I published a world with the name "materials.world"
    When I upload a material named "Stone"
    Then the material should appear in the materials list for the world
    When I delete the material
    Then the material delete response should be successful
