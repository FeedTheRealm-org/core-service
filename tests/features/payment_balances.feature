Feature: Payment Balances
  As a player
  I want to view my balances and gem packs
  So that I can manage purchases

  Background:
    Given I have logged in with email "payments_player@example.com" and password "Password123"

  Scenario: AC-1 Get gem balance
    When I request my gem balance
    Then I should receive my gem balance

  Scenario: AC-2 Get creator balance
    When I request my creator balance
    Then I should receive my creator balance

  Scenario: AC-3 List gem packs
    When I request gem packs
    Then I should receive the gem packs list
