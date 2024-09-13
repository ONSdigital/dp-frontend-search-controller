Feature: Previous Releases

#  Scenario: GET /previousreleases and checking for zero results
#    Given there is a Search API that gives a successful response and returns 0 results
#    And get page data request to zebedee is successful
#    And the search controller is running
#    When I navigate to "/economy/previousreleases"
#    Then the page should have the following content
#      """
#      {
#        "#main h1": "Previous releases for",
#        ".search__count h2": "0 results"
#      }
#      """

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

#  Scenario: GET /previousreleases and checking for thirty-six results
#    Given there is a Search API that gives a successful response and returns 36 results
#    And the search controller is running
#    When I navigate to "/businessindustryandtrade/changestobusiness/mergersandacquisitions/bulletins/mergersandacquisitionsinvolvingukcompanies/previousreleases"
#    Then the page should have the following content
#      """
#      {
#        "#main h1": "Previous releases for Mergers and acquisitions involving UK companies",
#        ".search__summary__count h2": "36 results "
#      }
#      """

#  Scenario: GET /previousreleases and checking for zero results
#    Given there is a Search API that gives a successful response and returns 0 results
#    And the search controller is running
#    When I navigate to "/someproduct/bulletins/somebulletin/previousreleases"
#    And the page should have the following content
#      """
#      {
#        "#main h1": "User requested data",
#        ".search__count h2": "0 results"
#      }
#      """

#  Scenario: GET /previousreleases and checking for one result
#    Given there is a Previous Releases requests that calls the Search API and returns 1 results
#    And the search controller is running
#    When I navigate to "/someproduct/bulletins/somebulletin/previousreleases"
#    Then the page should have the following content
#    """
#        {
#            ".search__summary": "Search summary"
#        }
#    """
