@TimeseriesTool
Feature: Timeseries Tool

  Scenario: GET /timeseriestool and checking for zero results
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/timeseriestool"
    And the page should have the following content
    """
        {
            "#main h1": "Time series explorer",
            ".search__count h2": "0 results"
        }
    """
    And element "#before-date" should be visible
    And element "#after-date" should be visible

  Scenario: GET /timeseriestool and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
    And the search controller is running
    When I navigate to "/timeseriestool"
    And the page should have the following content
    """
        {
            "#main h1": "Time series explorer",
            ".search__count h2": "1 result",
            ".ons-document-list__item-attribute:nth-child(2)": "Series ID: AA0",
            ".ons-document-list__item-attribute:nth-child(3)": "Dataset ID: DD0"
        }
    """

  Scenario: GET /timeseriestool and check invalid params - page
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/timeseriestool?page=5000000"
    Then the page should have the following content
    """
        {
            "h2#error-summary-title": "There is a problem with this page"
        }
    """

  Scenario: GET /timeseriestool and check invalid params - date
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/timeseriestool?after-month=13&after-year=2024"
    Then the page should have the following content
    """
        {
            "h2#error-summary-title": "There is a problem with this page"
        }
    """
