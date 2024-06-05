Feature: Form submission with date fields

    Background:
        Given there is a Search API that gives a successful response and returns 10 results
        And the search controller is running
        When I navigate to "/alladhocs"

    Scenario: Submitting the form with valid date fields
        Then I fill in "#after-date-day" with "01"
        And I fill in "#after-date-month" with "01"
        And I fill in "#after-date-year" with "2022"
        And I fill in "#before-date-day" with "31"
        And I fill in "#before-date-month" with "12"
        And I fill in "#before-date-year" with "2022"
        And I click the "#search-filter > button" button
        Then I wait 1 seconds
        Then element "#error-summary-title" should not be visible

    Scenario Outline: Submitting the form with invalid date fields
        Then I fill in "#after-date-day" with "<day>"
        And I fill in "#after-date-month" with "<month>"
        And I fill in "#after-date-year" with "<year>"
        And I click the "#search-filter > button" button
        Then I wait 1 seconds
        Then element "#error-summary-title" should be visible

        Examples:
            | day | month | year |
            | 32  | 01    | 2022 |
            | 01  | 13    | 2022 |
            | 01  | 01    | 0000 |
            | 01  | 01    |      |

    Scenario: Submitting the form with after date that's after before date
        Then I fill in "#after-date-day" with "01"
        And I fill in "#after-date-month" with "01"
        And I fill in "#after-date-year" with "2022"
        And I fill in "#before-date-day" with "31"
        And I fill in "#before-date-month" with "12"
        And I fill in "#before-date-year" with "2021"
        And I click the "#search-filter > button" button
        Then I wait 1 seconds
        Then element "#error-summary-title" should be visible
