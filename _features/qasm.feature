Feature:
    In order to run openqasm
    As an API User

    Scenario: should run bell.qasm
        Given I set upload file "testdata/bell.qasm"
        When I send "POST" request to "/"
        Then the response code should be 200
        Then the response should match json:
            """
            {
                "filename": "bell.qasm",
                "content": "OPENQASM 3.0;\n\ngate h q { U(pi/2.0, 0, pi) q; }\ngate x q { U(pi, 0, pi) q; }\ngate cx c, t { ctrl @ x c, t; }\n\nqubit[2] q;\nreset q;\n\nh q[0];\ncx q[0], q[1];\n",
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
