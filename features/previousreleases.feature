@previousReleases
Feature: Previous Releases

  Scenario: GET /previousreleases and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
    And get page data request to zebedee for "/economy/latest" returns a page of type "article" with status 200
    And get breadcrumb request to zebedee for "/economy/latest" returns breadcrumbs 
    And the search controller is running
    When I navigate to "/economy/previousreleases"
    Then the page should have the following content
    """
    {
      "#main h1": "Labour Market statistics",
      ".ons-pagination__position": "Page 1 of 1",
      ".ons-breadcrumb": "Home  Economy  Gross Domestic Product (GDP)  Labour Market statistics"
    }
    """

  Scenario: GET /previousreleases and checking for many results
    Given there is a Search API that gives a successful response and returns 3 results
    And get page data request to zebedee for "/economy/latest" returns a page of type "bulletin" with status 200
    And get breadcrumb request to zebedee for "/economy/latest" returns breadcrumbs
    And the search controller is running
    When I navigate to "/economy/previousreleases"
    Then the page should have the following content
      """
      {
        "#main h1": "Labour Market statistics",
        ".ons-pagination__position": "Page 1 of 1",
        ".ons-breadcrumb": "Home  Economy  Gross Domestic Product (GDP)  Labour Market statistics"
      }
      """

  Scenario: GET /previousreleases and breadcrumb request errors
  Given there is a Search API that gives a successful response and returns 3 results
  And get page data request to zebedee for "/economy/latest" returns a page of type "bulletin" with status 200
  And get breadcrumb request to zebedee for "/economy/latest" fails
  And the search controller is running
  When I navigate to "/economy/previousreleases"
  Then the page should have the following content
    """
    {
      "#main h1": "Labour Market statistics",
      ".ons-pagination__position": "Page 1 of 1",
      ".ons-breadcrumb": "Labour Market statistics "
    }
    """

  Scenario: GET /previousreleases and it is not a page of type article, bulletin, or compendium_landing_page
    Given get page data request to zebedee for "/economy/latest" returns a page of type "taxonomy_landing_page" with status 200
    And the search controller is running
    When I navigate to "/economy/previousreleases"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

  Scenario: GET /previousreleases and the latest release page does not exist
    Given get page data request to zebedee for "/economy/latest" does not find the page
    And the search controller is running
    When I navigate to "/economy/previousreleases"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """
