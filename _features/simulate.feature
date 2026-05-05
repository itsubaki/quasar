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
                "states": [
                    {
                        "probability": 0.5,
                        "amplitude": {
                            "real": 0.707107
                        },
                        "binaryString": [
                            "00"
                        ]
                    },
                    {
                        "probability": 0.5,
                        "amplitude": {
                            "real": 0.707107
                        },
                        "binaryString": [
                            "11"
                        ]
                    }
                ]
            }
            """

    Scenario: should run qft.qasm
        Given I set file "testdata/qft.qasm"
        Given I set "content-type" header with "application/json"
        Given I set request body:
            """
            {
                "code": "{{file:testdata/qft.qasm}}"
            }
            """
        When I send "POST" request to "/quasar.v1.QuasarService/Simulate"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "states": [
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "real": 0.353553
                        },
                        "binaryString": [
                            "000"
                        ]
                    },
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "real": 0.25,
                            "imag": 0.25
                        },
                        "binaryString": [
                            "001"
                        ]
                    },
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "imag": 0.353553
                        },
                        "binaryString": [
                            "010"
                        ]
                    },
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "real": -0.25,
                            "imag": 0.25
                        },
                        "binaryString": [
                            "011"
                        ]
                    },
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "real": -0.353553
                        },
                        "binaryString": [
                            "100"
                        ]
                    },
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "real": -0.25,
                            "imag": -0.25
                        },
                        "binaryString": [
                            "101"
                        ]
                    },
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "imag": -0.353553
                        },
                        "binaryString": [
                            "110"
                        ]
                    },
                    {
                        "probability": 0.125,
                        "amplitude": {
                            "real": 0.25,
                            "imag": -0.25
                        },
                        "binaryString": [
                            "111"
                        ]
                    }
                ]
            }
            """
