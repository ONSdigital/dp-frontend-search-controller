@AggregatedData
Feature: Form submission with date fields

  Background:
    Given there is a Search API that gives a successful response and returns 10 results
    And the search controller is running
    And I navigate to "/alladhocs"

  Scenario: Submitting the form with valid date fields
    Given I fill in input element "#after-date-day" with value "01"
    And I fill in input element "#after-date-month" with value "01"
    And I fill in input element "#after-date-year" with value "2022"
    And I fill in input element "#before-date-day" with value "31"
    And I fill in input element "#before-date-month" with value "12"
    And I fill in input element "#before-date-year" with value "2022"
    When I click the "#search-filter > button" element
    Then element "#error-summary-title" should not be visible

  Scenario Outline: Submitting the form with invalid after-date fields
    Given I fill in input element "#after-date-day" with value "<day>"
    And I fill in input element "#after-date-month" with value "<month>"
    And I fill in input element "#after-date-year" with value "<year>"
    When I click the "#search-filter > button" element
    Then element "#error-summary-title" should be visible
    And element "#after-date-error" should be visible

    Examples:
      | day | month | year |
      | 32  | 01    | 2022 |
      | 01  | 13    | 2022 |
      | 01  | 01    | 0000 |
      | 31  | 09    | 2022 |
      | 01  | 01    |      |

  Scenario Outline: Submitting the form with invalid before-date fields
    Given I fill in input element "#before-date-day" with value "<day>"
    And I fill in input element "#before-date-month" with value "<month>"
    And I fill in input element "#before-date-year" with value "<year>"
    When I click the "#search-filter > button" element
    Then element "#error-summary-title" should be visible
    And element "#before-date-error" should be visible

    Examples:
      | day | month | year |
      | 32  | 01    | 2022 |
      | 01  | 13    | 2022 |
      | 01  | 01    | 0000 |
      | 31  | 09    | 2022 |
      | 01  | 01    |      |

  Scenario: Submitting the form with after-date that's after before-date
    Given I fill in input element "#after-date-day" with value "01"
    And I fill in input element "#after-date-month" with value "01"
    And I fill in input element "#after-date-year" with value "2022"
    And I fill in input element "#before-date-day" with value "31"
    And I fill in input element "#before-date-month" with value "12"
    And I fill in input element "#before-date-year" with value "2021"
    When I click the "#search-filter > button" element
    Then element "#error-summary-title" should be visible
    And element "#after-date-error" should not be visible
    And element "#before-date-error" should be visible
