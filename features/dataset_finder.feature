Feature: Dataset Finder
  Scenario: GET / and checking the response status 200
    When I navigate to "/census/find-a-dataset"
    Then the improve this page banner should be visible
    # breadcrumb
    And the page should have the following content
    """
        {
            ".ons-breadcrumb > ol > li:nth-child(1) > a": "Home",
            ".ons-breadcrumb > ol > li:nth-child(2) > a": "Census",
            ".ons-breadcrumb > ol > li:nth-child(3)": "Find census data" 
        }
    """