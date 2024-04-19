Feature: Search
  Scenario: GET /search and checking for zero results
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/search?q=test+query"
    And the page should have the following content
    """
        {
            "#main h1": "Search results for test query",
            ".search__summary__count": "0 results"
        }
    """
  Scenario: GET /search and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
    And the search controller is running
    When I navigate to "/search?q=test+query"
    And the page should have the following content
    """
        {
            "#main h1": "Search results for test query",
            ".search__count h2": "1 result",
            ".ons-pagination__position": "Page 1 of 1"
        }
    """

  Scenario: GET /search and checking for 10 results
    Given there is a Search API that gives a successful response and returns 10 results
    And the search controller is running
    When I navigate to "/search?q=test+query"
    And the page should have the following content
    """
        {
            "#main h1": "Search results for test query",
            ".search__count h2": "10 results",
            ".ons-pagination__position": "Page 1 of 1"
        }
    """
  Scenario: GET /search and checking for 11 results
    Given there is a Search API that gives a successful response and returns 11 results
    And the search controller is running
    When I navigate to "/search?q=test+query"
    And the page should have the following content
    """
        {
            "#main h1": "Search results for test query",
            ".search__count h2": "11 results",
            ".ons-pagination__position": "Page 1 of 2"
        }
    """
  Scenario: GET /search with no query
    Given the search controller is running
    When I navigate to "/search"
    And the page should have the following content
    """
        {
            ".search__summary__generic": "Enter a search term"
        }
    """
