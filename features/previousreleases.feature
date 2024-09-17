Feature: Previous Releases

  Scenario: GET /previousreleases and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
    And get page data request to zebedee is successful
    And the search controller is running
    When I navigate to "/economy/previousreleases"
    Then the page should have the following content
    """
    {
      "#main h1": "Previous releases for",
      ".search__count h2": "1 result",
      ".ons-pagination__position": "Page 1 of 1"
    }
    """

  Scenario: GET /previousreleases and checking for many results
    Given there is a Search API that gives a successful response and returns 3 results
    And get page data request to zebedee is successful
    And the search controller is running
    When I navigate to "/economy/previousreleases"
    Then the page should have the following content
      """
      {
        "#main h1": "Previous releases for",
        ".search__count h2": "3 results",
        ".ons-pagination__position": "Page 1 of 1"
      }
      """

  Scenario: GET /previousreleases and it is the wrong page type
    Given get page data request to zebedee finds a wrong page type
    And the search controller is running
    When I navigate to "/economy/previousreleases"
    Then the page should have the following content
    """
        {
            "#main h1": "Page not found"
        }
    """

#  Scenario: GET /previousreleases and checking for zero results even though zebedee finds the right page type
#    Given there is a Search API that gives a successful response and returns 0 results
#    And get page data request to zebedee is successful
#    And the search controller is running
#    When I navigate to "/economy/previousreleases"
#    Then the page should have the following content
#    """
#        {
#            "#main h1": "No releases found"
#        }
#    """
