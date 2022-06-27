Feature:
    In order to run quantum algorithm with openqasm
    As an API User

    Scenario: should run bell.qasm
        Given I set upload file "_testdata/bell.qasm"
        When I send "POST" request to "/"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "state": [
                    {
                        "amplitude": {
                            "real": 0.7071067811865476,
                            "imag": 0
                        },
                        "probability": 0.5000000000000001,
                        "int": [
                            0
                        ],
                        "binary_string": [
                            "00"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": 0.7071067811865475,
                            "imag": 0
                        },
                        "probability": 0.4999999999999999,
                        "int": [
                            3
                        ],
                        "binary_string": [
                            "11"
                        ]
                    }
                ]
            }
            """
