Feature: Aggregated Data Pages

  Scenario: GET /alladhocs and checking for zero results
    Given there is a Search API that gives a successful response and returns 0 results
    When I navigate to "/alladhocs"
    And the page should have the following content
    """
        {
            "#main h1": "User requested data",
            ".search__count h2": "0 results"
        }
    """
  Scenario: GET /alladhocs and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
    When I navigate to "/alladhocs"
    And the page should have the following content
    """
        {
            "#main h1": "User requested data",
            ".search__count h2": "1 result",
            ".ons-pagination__position": "Page 1 of 1"
        }
    """
  Scenario: GET /alladhocs and checking for 10 results
    Given there is a Search API that gives a successful response and returns 10 results
    When I navigate to "/alladhocs"
    And the page should have the following content
    """
        {
            "#main h1": "User requested data",
            ".search__count h2": "10 results",
            ".ons-pagination__position": "Page 1 of 1"
        }
    """
  Scenario: GET /alladhocs and checking for 11 results
    Given there is a Search API that gives a successful response and returns 11 results
    When I navigate to "/alladhocs"
    And the page should have the following content
    """
        {
            "#main h1": "User requested data",
            ".search__count h2": "11 results",
            ".ons-pagination__position": "Page 1 of 2"
        }
    """
  Scenario: GET /alladhocs and check default sort
    Given there is a Search API that gives a successful response and returns 10 results
    When I navigate to "/alladhocs"
    Then input element "#sort" has value "release_date"

  Scenario: GET /alladhocs and check param driven sort
    Given there is a Search API that gives a successful response and returns 10 results
    When I navigate to "/alladhocs?sort=relevance"
    Then input element "#sort" has value "relevance"
