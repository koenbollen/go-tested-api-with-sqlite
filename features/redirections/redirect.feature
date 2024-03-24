Feature: Redirect http GET requests to the appropriate URL

  I want to be able to redirect http GET requests to the appropriate URL
  So that I can ensure that users are directed to the correct destination

  Scenario: Redirect the client to a stored url by key
    Given the follow "redirection" record exist:
      | key        | test                 |
      | url        | http://example.com   |
    When the client does a GET request to "/test"
    Then the response code should be 302 (Found)
    And the response header "Location" should be "http://example.com"
    And the response body should be the following "text/html":
      """html
      <script type=\"text/javascript\">window.location = "http://example.com";</script>
      """

  Scenario: Fail to redirect to a non-existing key
    When the client does a GET request to "/does-not-exists"
    Then the response code should be 404 (Not Found)
