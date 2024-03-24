Feature: Redirections

  As an example route this project allows clients to create and get redirections.

  Scenario: Create a redirections
    When the client does a POST request to "/redirections" with the following data:
      """json
      {
        "key": "rickroll",
        "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
      }
      """
    Then the response code should be 200 (Created)
    And this "redirection" record exists:
      | key        | rickroll                                    |
      | url        | https://www.youtube.com/watch?v=dQw4w9WgXcQ |
      | created_at | 2009-11-10T23:00:00Z                        |
      | updated_at | 2009-11-10T23:00:00Z                        |

  Scenario: Fail to create redirection without an url
    When the client does a POST request to "/redirections" with the following data:
      """json
      {
        "key": "test"
      }
      """
    Then the response code should be 400 (Bad Request)
    And the response body should be the following "application/json":
      """json
      {
        "error": "key and url are required"
      }
      """

  Scenario: Delete a redirection by key
    Given the follow "redirection" record exist:
      | key        | test                 |
      | url        | http://example.com   |
      | created_at | 2009-11-10T23:00:00Z |
      | updated_at | 2009-11-10T23:00:00Z |
    When the client does a DELETE request to "/redirections/test"
    Then the response code should be 204 (No Content)
    And no "redirection" record exists with id "test"

