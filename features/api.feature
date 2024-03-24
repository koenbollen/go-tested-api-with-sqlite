Feature: Basic API

  Scenario: Healthcheck
    When the client does a GET request to "/health"
    Then the response code should be 200 (OK)
    And the response body should be empty
