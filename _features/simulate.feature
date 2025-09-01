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
                        "amplitude": {
                            "real": 0.707106
                        },
                        "probability": 0.5,
                        "int": [
                            "0"
                        ],
                        "binaryString": [
                            "00"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": 0.707106
                        },
                        "probability": 0.499999,
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
                "state": [
                    {
                        "amplitude": {
                            "real": 0.353553
                        },
                        "probability": 0.125,
                        "int": [
                            "0"
                        ],
                        "binaryString": [
                            "000"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": 0.25,
                            "imag": 0.249999
                        },
                        "probability": 0.125,
                        "int": [
                            "1"
                        ],
                        "binaryString": [
                            "001"
                        ]
                    },
                    {
                        "amplitude": {
                            "imag": 0.353553
                        },
                        "probability": 0.124999,
                        "int": [
                            "2"
                        ],
                        "binaryString": [
                            "010"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": -0.249999,
                            "imag": 0.249999
                        },
                        "probability": 0.124999,
                        "int": [
                            "3"
                        ],
                        "binaryString": [
                            "011"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": -0.353553
                        },
                        "probability": 0.125,
                        "int": [
                            "4"
                        ],
                        "binaryString": [
                            "100"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": -0.25,
                            "imag": -0.249999
                        },
                        "probability": 0.125,
                        "int": [
                            "5"
                        ],
                        "binaryString": [
                            "101"
                        ]
                    },
                    {
                        "amplitude": {
                            "imag": -0.353553
                        },
                        "probability": 0.125,
                        "int": [
                            "6"
                        ],
                        "binaryString": [
                            "110"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": 0.249999,
                            "imag": -0.25
                        },
                        "probability": 0.125,
                        "int": [
                            "7"
                        ],
                        "binaryString": [
                            "111"
                        ]
                    }
                ]
            }
            """
