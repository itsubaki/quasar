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
                            "real": 0.7071067811865476
                        },
                        "probability": 0.5000000000000001,
                        "int": [
                            "0"
                        ],
                        "binaryString": [
                            "00"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": 0.7071067811865475
                        },
                        "probability": 0.4999999999999999,
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
                            "real": 0.3535533905932738
                        },
                        "probability": 0.12500000000000003,
                        "int": [
                            "0"
                        ],
                        "binaryString": [
                            "000"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": 0.2500000000000001,
                            "imag": 0.24999999999999992
                        },
                        "probability": 0.12500000000000003,
                        "int": [
                            "1"
                        ],
                        "binaryString": [
                            "001"
                        ]
                    },
                    {
                        "amplitude": {
                            "imag": 0.35355339059327373
                        },
                        "probability": 0.12499999999999997,
                        "int": [
                            "2"
                        ],
                        "binaryString": [
                            "010"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": -0.24999999999999986,
                            "imag": 0.24999999999999997
                        },
                        "probability": 0.1249999999999999,
                        "int": [
                            "3"
                        ],
                        "binaryString": [
                            "011"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": -0.35355339059327384
                        },
                        "probability": 0.12500000000000006,
                        "int": [
                            "4"
                        ],
                        "binaryString": [
                            "100"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": -0.25000000000000017,
                            "imag": -0.24999999999999992
                        },
                        "probability": 0.12500000000000003,
                        "int": [
                            "5"
                        ],
                        "binaryString": [
                            "101"
                        ]
                    },
                    {
                        "amplitude": {
                            "imag": -0.3535533905932738
                        },
                        "probability": 0.12500000000000003,
                        "int": [
                            "6"
                        ],
                        "binaryString": [
                            "110"
                        ]
                    },
                    {
                        "amplitude": {
                            "real": 0.2499999999999999,
                            "imag": -0.2500000000000001
                        },
                        "probability": 0.12500000000000003,
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
