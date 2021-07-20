Feature: Healthcheck endpoint should inform the health of service

    Scenario: Returning a OK (200) status when health endpoint called  
        Given all the downstream services are healthy
        And I wait "2" seconds for the healthcheck to be available
        When I GET "/health"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json; charset=utf-8"
        And I should receive the following health JSON response:
        """
            {
                "status":"OK",
                "version":{
                    "build_time":"0001-01-01T00:00:00Z",
                    "git_commit":"componentGitCommit",
                    "language":"go",
                    "language_version":"go1.16.5",
                    "version":"componentVersion"
                },
                "uptime":1234,
                "start_time":"0001-01-01T00:00:00Z",
                "checks":[
                    {
                        "name":"frontend renderer",
                        "status":"OK",
                        "status_code":200,
                        "message":"renderer is ok",
                        "last_checked":"0001-01-01T00:00:00Z",
                        "last_success":"0001-01-01T00:00:00Z",
                        "last_failure": null
                    },
                    {
                        "name":"API router",
                        "status":"OK",
                        "status_code":200,
                        "message":"api-router is ok",
                        "last_checked":"0001-01-01T00:00:00Z",
                        "last_success":"0001-01-01T00:00:00Z",
                        "last_failure": null
                    }
                ]
            }
        """

    Scenario: Returning a WARNING (429) status when one downstream service is warning  
        Given one the downstream services is warning
        And I wait "2" seconds for the healthcheck to be available
        When I GET "/health"
        Then the HTTP status code should be "429"
        And the response header "Content-Type" should be "application/json; charset=utf-8"
        And I should receive the following health JSON response:
        """
            {
                "status": "WARNING",
                "version": {
                    "build_time": "0001-01-01T00:00:00Z",
                    "git_commit": "componentGitCommit",
                    "language": "go",
                    "language_version": "go1.16.5",
                    "version": "componentVersion"
                },
                "uptime": 1234,
                "start_time": "0001-01-01T00:00:00Z",
                "checks": [
                    {
                        "name": "frontend renderer",
                        "status": "OK",
                        "status_code": 200,
                        "message": "renderer is ok",
                        "last_checked": "0001-01-01T00:00:00Z",
                        "last_success": "0001-01-01T00:00:00Z",
                        "last_failure": null
                    },
                    {
                        "name": "API router",
                        "status": "WARNING",
                        "status_code": 429,
                        "message": "api-router is degraded, but at least partially functioning",
                        "last_checked": "0001-01-01T00:00:00Z",
                        "last_success": null,
                        "last_failure": "0001-01-01T00:00:00Z"
                    }
                ]
            }
        """

    Scenario: Returning a WARNING (429) status when one downstream service is critical and critical timeout has not expired  
        Given one the downstream services is failing
        And I wait "2" seconds for the healthcheck to be available
        When I GET "/health"
        Then the HTTP status code should be "429"
        And the response header "Content-Type" should be "application/json; charset=utf-8"
        And I should receive the following health JSON response:
        """
            {
                "status": "WARNING",
                "version": {
                    "build_time": "0001-01-01T00:00:00Z",
                    "git_commit": "componentGitCommit",
                    "language": "go",
                    "language_version": "go1.16.5",
                    "version": "componentVersion"
                },
                "uptime": 1234,
                "start_time": "0001-01-01T00:00:00Z",
                "checks": [
                    {
                        "name": "frontend renderer",
                        "status": "OK",
                        "status_code": 200,
                        "message": "renderer is ok",
                        "last_checked": "0001-01-01T00:00:00Z",
                        "last_success": "0001-01-01T00:00:00Z",
                        "last_failure": null
                    },
                    {
                        "name": "API router",
                        "status": "CRITICAL",
                        "status_code": 500,
                        "message": "api-router functionality is unavailable or non-functioning",
                        "last_checked": "0001-01-01T00:00:00Z",
                        "last_success": null,
                        "last_failure": "0001-01-01T00:00:00Z"
                    }
                ]
            }
        """

    Scenario: Returning a CRITICAL (500) status when health endpoint called
        Given one the downstream services is failing
        And I wait "2" seconds for the healthcheck to be available
        When I GET "/health"
        And I wait "3" seconds to pass the critical timeout
        And I GET "/health"
        Then the HTTP status code should be "500"
        And the response header "Content-Type" should be "application/json; charset=utf-8"
        And I should receive the following health JSON response:
        """
            {
                "status": "CRITICAL",
                "version": {
                    "build_time": "0001-01-01T00:00:00Z",
                    "git_commit": "componentGitCommit",
                    "language": "go",
                    "language_version": "go1.16.5",
                    "version": "componentVersion"
                },
                "uptime": 1234,
                "start_time": "0001-01-01T00:00:00Z",
                "checks": [
                    {
                        "name": "frontend renderer",
                        "status": "OK",
                        "status_code": 200,
                        "message": "renderer is ok",
                        "last_checked": "0001-01-01T00:00:00Z",
                        "last_success": "0001-01-01T00:00:00Z",
                        "last_failure": null
                    },
                    {
                        "name": "API router",
                        "status": "CRITICAL",
                        "status_code": 500,
                        "message": "api-router functionality is unavailable or non-functioning",
                        "last_checked": "0001-01-01T00:00:00Z",
                        "last_success": null,
                        "last_failure": "0001-01-01T00:00:00Z"
                    }
                ]
            }
        """