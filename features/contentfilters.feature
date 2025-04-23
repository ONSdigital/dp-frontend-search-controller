Feature: Content type filter checkboxes on /search

  Background:
    Given there is a Search API that gives a successful response and returns 10 results
    And the search controller is running
    And I navigate to "/search?q=test+query"

  Scenario: Publications filter checkbox is visible and subtypes are hidden
    Then input element "#group-0" has value "publications"
    When I click the "#group-0" element
    Then element "#bulletin" should not be visible
    Then element "#article" should not be visible
    Then element "#compendiums" should not be visible
    Then element "#statistical_article" should not be visible

  Scenario: Data filter checkbox is visible and subtypes are visible
    Then element "#group-1" should be visible
    When I click the "#group-1" element
    Then element "#time_series" should be visible
    Then element "#datasets" should be visible
    Then element "#user_requested_data" should be visible

  Scenario: Other filter checkbox is visible and subtypes are visible
    Then element "#group-2" should be visible
    When I click the "#group-2" element
    Then element "#methodology" should be visible
    Then element "#corporate_information" should be visible
