Feature:
    In order to factor numbers
    As an API User

    Scenario: should factor the number 15
        When I send "GET" request to "/shor/15?a=7&seed=1"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "N": 15,
                "a": 7,
                "m": "0.010",
                "p": 3,
                "q": 5,
                "s/r": "1/4",
                "seed": 1,
                "t": 3
            }
            """
