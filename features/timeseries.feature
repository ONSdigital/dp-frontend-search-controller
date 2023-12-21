Feature: Timeseries Tool

  Scenario: GET /timeseriestool and checking for zero results
    Given there is a Search API that gives a successful response and returns 0 results
    When I navigate to "/timeseriestool"
    And the page should have the following content
    """
        {
            "#main h1": "Time series explorer",
            ".search__count h2": "0 results"
        }
    """
  Scenario: GET /alladtimeseriestoolhocs and checking for one result
    Given there is a Search API that gives a successful response and returns 1 results
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
