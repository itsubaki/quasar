Feature:
    In order to run quantum algorithm with openqasm
    As an API User

    Scenario: should run bell.qasm
        Given I set file "testdata/bell.qasm"
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "code": "{{file:testdata/bell.qasm}}"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Simulate"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "state": [
                    {
                        "probability": 0.5000000000000001,
                        "amplitude": {
                            "real": 0.7071067811865476
                        },
                        "int": [
                            "0"
                        ],
                        "binaryString": [
                            "00"
                        ]
                    },
                    {
                        "probability": 0.5000000000000001,
                        "amplitude": {
                            "real": 0.7071067811865476
                        },
                        "int": [
                            "3"
                        ],
                        "binaryString": [
                            "11"
                        ]
                    }
                ]
            }
            """
