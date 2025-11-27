
@A11y
# TODO: extract site furniture a11y tests to dis-design-system-go when ready 
Feature: A11y checks on /search

    Background:
        Given there is a Search API that gives a successful response and returns 10 results
        And the search controller is running
        And I navigate to "/search?q=test+query"

    Scenario: Page is accessible in desktop view
        Then the page should be accessible

    Scenario: Page is accessible in mobile view
        When I set the viewport to mobile
        And I click the "#search-toggle" element
        Then the page should be accessible

    Scenario: Page is accessible in tablet view
        When I set the viewport to tablet
        Then the page should be accessible
