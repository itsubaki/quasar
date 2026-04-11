Feature:
    In order to validate openqasm code
    As an API User

    Scenario: should validate bell.qasm
        Given I set file "testdata/bell.qasm"
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "code": "{{file:testdata/bell.qasm}}"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Validate"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "valid": true
            }
            """

    Scenario: should validate invalid code
        Given I set file "testdata/bell.qasm"
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "code": "qubit[ q;"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Validate"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "line": 1,
                "column": 8,
                "message": "mismatched input ';' expecting ']'"
            }
            """
