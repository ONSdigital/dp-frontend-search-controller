@RelatedData
Feature: Related Data

  Scenario: GET /relateddata and checking for one result
    Given there is a Search API that gives a successful Search URIs response and returns 1 results
    And get page data request to zebedee for "/economy/article" returns a page of type "article" with status 200
    And get breadcrumb request to zebedee for "/economy/article" returns breadcrumbs 
    And the search controller is running
    When I navigate to "/economy/article/relateddata"
    Then the page should have the following content
    """
    {
      "#main h1": "All data related to Labour Market statistics: March 2024",
      ".ons-pagination__position": "Page 1 of 1",
      ".ons-breadcrumb": "Home  Economy  Gross Domestic Product (GDP)  Labour Market statistics"
    }
    """

  Scenario: GET /previousreleases and checking for many results
    Given there is a Search API that gives a successful Search URIs response and returns 3 results
    And get page data request to zebedee for "/economy/article" returns a page of type "bulletin" with status 200
    And get breadcrumb request to zebedee for "/economy/article" returns breadcrumbs
    And the search controller is running
    When I navigate to "/economy/article/relateddata"
    Then the page should have the following content
      """
      {
        "#main h1": "All data related to Labour Market statistics: March 2024",
        ".ons-pagination__position": "Page 1 of 1",
        ".ons-breadcrumb": "Home  Economy  Gross Domestic Product (GDP)  Labour Market statistics"
      }
      """

