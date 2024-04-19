Feature: Aggregated Data Pages

  Scenario: GET /alladhocs and checking for zero results
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/alladhocs"
    And the page should have the following content
      """
      {
        "#main h1": "User requested data",
        ".search__count h2": "0 results"
      }
      """

  Scenario: GET /alladhocs and date fieldsets are displayed
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/alladhocs"
    Then element "#to-date-filters" should be visible
    Then element "#from-date-filters" should be visible

  Scenario: GET /alladhocs and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
    And the search controller is running
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
    And the search controller is running
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
    And the search controller is running
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
    And the search controller is running
    When I navigate to "/alladhocs"
    Then input element "#sort" has value "release_date"

  Scenario: GET /alladhocs and check param driven sort
    Given there is a Search API that gives a successful response and returns 10 results
    And the search controller is running
    When I navigate to "/alladhocs?sort=relevance"
    Then input element "#sort" has value "relevance"

  Scenario: GET topic pre-filtered page with matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" root topic
    And the search controller is running
    When I navigate to "/economy/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Publications related to economy",
            ".search__count h2": "10 results"
        }
    """

  Scenario: GET topic pre-filtered page with non-matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API returns no topics
    And the search controller is running
    When I navigate to "/testpath/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  Scenario: GET subtopic pre-filtered page with matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" root topic and the "environmental" subtopic
    And the search controller is running
    When I navigate to "/economy/environmental/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Publications related to environmental",
            ".search__count h2": "10 results"
        }
    """

  Scenario: GET subtopic pre-filtered page with non-matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" root topic and the "environmental" subtopic
    And the search controller is running
    When I navigate to "/economy/testtopic/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """
