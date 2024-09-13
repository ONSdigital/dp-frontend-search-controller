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
    Then element "#before-date" should be visible
    Then element "#after-date" should be visible

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

  Scenario: GET /alladhocs and check invalid params - page
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/alladhocs?page=5000000"
    Then the page should have the following content
    """
        {
            "h2#error-summary-title": "There is a problem with this page"
        }
    """

  Scenario: GET /alladhocs and check invalid params - page
    Given there is a Search API that gives a successful response and returns 500 results
    And the search controller is running
    When I navigate to "/alladhocs?page=51"
    Then the page should have the following content
    """
        {
            "h2#error-summary-title": "There is a problem with this page"
        }
    """

  Scenario: GET /alladhocs and check invalid params - date
    Given there is a Search API that gives a successful response and returns 0 results
    And the search controller is running
    When I navigate to "/alladhocs?after-month=13&after-year=2024"
    Then the page should have the following content
    """
        {
            "h2#error-summary-title": "There is a problem with this page"
        }
    """

  @topicPages
  Scenario: GET topic pre-filtered page with matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" topic
    And the search controller is running
    When I navigate to "/economy/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Publications related to economy",
            ".search__count h2": "10 results"
        }
    """

  @topicPages
  Scenario: GET topic pre-filtered page with non-matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" topic
    And the search controller is running
    When I navigate to "/testpath/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  @topicPages
  Scenario: GET subtopic pre-filtered page with matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" topic and the "environmentalaccounts" subtopic
    And the search controller is running
    When I navigate to "/economy/environmentalaccounts/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Publications related to environmental accounts",
            ".search__count h2": "10 results"
        }
    """

  @topicPages
  Scenario: GET subtopic pre-filtered page with non-matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" topic and the "environmentalaccounts" subtopic
    And the search controller is running
    When I navigate to "/economy/testtopic/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  @topicPages
  Scenario: GET subtopic pre-filtered page with wrong topic hierarchy
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "business" topic and the "environmentalaccounts" subtopic
    And the search controller is running
    When I navigate to "/business/environmentalaccounts/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  @topicPages
  Scenario: GET 3rd level subtopic pre-filtered page with matching topic and subtopic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" topic, the "governmentpublicsectorandtaxes" subtopic and "publicsectorfinance" thirdlevel subtopic
    And the search controller is running
    When I navigate to "/economy/governmentpublicsectorandtaxes/publicsectorfinance/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Publications related to public sector finance",
            ".search__count h2": "10 results"
        }
    """

  @topicPages
  Scenario: GET 3rd level subtopic pre-filtered page with non-matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" topic, the "governmentpublicsectorandtaxes" subtopic and "publicsectorfinance" thirdlevel subtopic
    And the search controller is running
    When I navigate to "/economy/governmentpublicsectorandtaxes/testtopic/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  @topicPages
  Scenario: GET 3rd level subtopic pre-filtered page with wrong topic hierarchy
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "business" topic, the "governmentpublicsectorandtaxes" subtopic and "publicsectorfinance" thirdlevel subtopic
    And the search controller is running
    When I navigate to "/business/governmentpublicsectorandtaxes/publicsectorfinance/publications"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  @agg_rss
  @topicPages
  Scenario: GET rss for subtopic pre-filtered page with matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" root topic and the "environmental" subtopic for requestQuery "rss"
    And the search controller is running
    When I navigate to "/economy/environmental/publications?rss"
    Then the page should have the following xml content
    """
      <?xml version="1.0" encoding="ISO-8859-1"?>
      <rss xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
      <channel>
        <title>Latest ONS releases matching (topic: /economy/environmental, type: data).</title>
        <link>http://dp.aws.onsdigital.uk/economy/environmental?rss</link>
        <description>Latest ONS releases</description>
        <category>/economy/environmental</category>
        <dc:subject>/economy/environmental</dc:subject>
      </channel>
      </rss>
    """
    And the response header "Content-Type" should contain "application/rss+xml; charset=ISO-8859-1"

  @agg_rss
  @topicPages
  Scenario: GET RSS subtopic pre-filtered page with non-matching topic
    Given there is a Search API that gives a successful response and returns 10 results
    And there is a Topic API that returns the "economy" root topic and the "environmental" subtopic for requestQuery "rss"
    And the search controller is running
    When I navigate to "/economy/testtopic/publications/rss"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """