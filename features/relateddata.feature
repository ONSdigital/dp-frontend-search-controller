@RelatedData
Feature: Related Data

  Scenario: GET /relateddata and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
    And get page data request to zebedee for "/employmentandlabourmarket/peopleinwork/article1" returns a page of type "article" with status 200
    And get breadcrumb request to zebedee for "/employmentandlabourmarket/peopleinwork/article1" returns breadcrumbs
    And the search controller is running
    When I navigate to "/employmentandlabourmarket/peopleinwork/article1/relateddata"
    Then the page should have the following content
    """
    {
      "#main h1": "All data related to Labour Market statistics: March 2024",
      ".ons-pagination__position": "Page 1 of 1",
      ".ons-document-list h2": "Test Bulletin 0",
      ".ons-breadcrumb": "Home  Economy  Gross Domestic Product (GDP)  Labour Market statistics"
    }
    """

  Scenario: GET /relateddata and checking for many results
    Given there is a Search API that gives a successful response and returns 3 results
    And get page data request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin1" returns a page of type "bulletin" with status 200
    And get breadcrumb request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin1" returns breadcrumbs
    And the search controller is running
    When I navigate to "/employmentandlabourmarket/peopleinwork/bulletin1/relateddata"
    Then the page should have the following content
      """
      {
        "#main h1": "All data related to Labour Market statistics: March 2024",
        ".ons-pagination__position": "Page 1 of 1",
        ".ons-document-list h2": "Test Bulletin 0",
        ".ons-breadcrumb": "Home  Economy  Gross Domestic Product (GDP)  Labour Market statistics"
      }
      """

  Scenario: GET /relateddata with a migration link
    Given there is a Search API that gives a successful response and returns 1 results
    And get page data request to zebedee for "/employmentandlabourmarket/bulletin1/latest" returns a page with migration link "/new-weekly-earnings"
    And the search controller is running
    When I GET "/employmentandlabourmarket/bulletin1/latest/relateddata"
    Then the HTTP status code should be "308"
    And the response header "Location" should be "/new-weekly-earnings/related-data"

  Scenario: GET /relateddata  and breadcrumb request errors
    Given there is a Search API that gives a successful response and returns 3 results
    And get page data request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin1" returns a page of type "bulletin" with status 200
    And get breadcrumb request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin1" fails
    And the search controller is running
    When I navigate to "/employmentandlabourmarket/peopleinwork/bulletin1/relateddata"
    Then the page should have the following content
    """
    {
      "#main h1": "All data related to Labour Market statistics: March 2024",
      ".ons-pagination__position": "Page 1 of 1",
      ".ons-document-list h2": "Test Bulletin 0",
      ".ons-breadcrumb": "Labour Market statistics "
    }
    """

  Scenario: GET /relateddata and it is not a page of type article, bulletin, or compendium_landing_page
    Given get page data request to zebedee for "/economy/latest" returns a page of type "taxonomy_landing_page" with status 200
    And the search controller is running
    When I navigate to "/economy/relateddata"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  Scenario: GET /relateddata and has no related data
    Given there is a Search API that gives a successful response and returns 0 results
    Given get page data request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin2" does not have related data
    And get breadcrumb request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin2" returns breadcrumbs
    And the search controller is running
    When I navigate to "/employmentandlabourmarket/peopleinwork/bulletin2/relateddata"
    Then the page should have the following content
    """
        {
            "#main h1": "All data related to Labour Market statistics: March 2024",
            "#no-results-text": "Sorry, there are no matching results."
        }
    """
    And element ".pagination" should not be visible

  Scenario: GET /relateddata and check invalid params - page
    Given there is a Search API that gives a successful response and returns 3 results
    And get page data request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin1" returns a page of type "bulletin" with status 200
    And get breadcrumb request to zebedee for "/employmentandlabourmarket/peopleinwork/bulletin1" returns breadcrumbs
    And the search controller is running
    When I navigate to "/employmentandlabourmarket/peopleinwork/bulletin1/relateddata?page=5000000"
    Then the page should have the following content
      """
      {
        "h2#error-summary-title": "There is a problem with this page"
      }
      """

