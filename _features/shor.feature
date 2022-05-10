Feature:
    In order to factoring numbers
    As an API User

    Scenario: should factoring 15
        When I send "GET" request to "/shor/15?a=7&seed=1"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "N": 15,
                "a": 7,
                "m": "0.110",
                "p": 3,
                "q": 5,
                "s/r": "3/4",
                "t": 3
            }
            """
