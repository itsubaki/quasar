Feature:
    In order to share quantum algorithm with openqasm
    As an API User

    Scenario: should save bell.qasm
        Given I set file "testdata/bell.qasm"
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "code": "{{file:testdata/bell.qasm}}"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Save"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "id": "AMOYU8a1VLEfWjqf",
                "createdAt": "@string@"
            }
            """

    Scenario: should load bell.qasm
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "id": "AMOYU8a1VLEfWjqf"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Load"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "id": "AMOYU8a1VLEfWjqf",
                "code": "@string@",
                "createdAt": "@string@"
            }
            """

