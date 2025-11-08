Feature:
    In order to share a quantum algorithm with openqasm
    As an API User

    Scenario: should share bell.qasm
        Given I set file "testdata/bell.qasm"
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "code": "{{file:testdata/bell.qasm}}"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Share"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "id": "AMOYU8a1VLEfWjqf",
                "createdAt": "@string@"
            }
            """

    Scenario: should edit bell.qasm
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "id": "AMOYU8a1VLEfWjqf"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Edit"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "id": "AMOYU8a1VLEfWjqf",
                "code": "@string@",
                "createdAt": "@string@"
            }
            """

