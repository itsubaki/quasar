Feature:
    In order to factor numbers
    As an API User

    Scenario: should factorize 15
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "n": "15",
                "a": "7",
                "t": "3",
                "seed": "1"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Factorize"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "n": "15",
                "a": "7",
                "t": "3",
                "seed": "1",
                "m": "0.010",
                "sr": "1/4",
                "p": "3",
                "q": "5"
            }
            """
