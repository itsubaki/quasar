# quasar

Quantum Computation Simulator as a Service

## Deployment

```shell
$ export PROJECT_ID=YOUR_GOOGLE_CLOUD_PROJECT_ID
$ gcloud builds submit --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/quasar
$ gcloud run deploy quasar --project ${PROJECT_ID} --image gcr.io/${PROJECT_ID}/quasar --set-env-vars=GOOGLE_CLOUD_PROJECT=${PROJECT_ID}
```

## Example

```shell
$ cat testdata/bell.qasm
OPENQASM 3.0;

gate h q { U(pi/2.0, 0, pi) q; }
gate x q { U(pi, 0, pi) q; }
gate cx c, t { ctrl @ x c, t; }

qubit[2] q;
reset q;

h q[0];
cx q[0], q[1];

$ curl -s -H "Authorization: Bearer $(gcloud auth print-identity-token)" -X POST -F file=@testdata/bell.qasm https://quasar-abcdefghij-an.a.run.app | jq .
{
  "trace_id": "4ec0662be027cb2904f39b19d197b27b",
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
```

```shell
$ curl -s -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://quasar-abcdefghij-an.a.run.app/shor/15 | jq .
{
  "N": 15,
  "a": 13,
  "m": "0.010",
  "p": 3,
  "q": 5,
  "s/r": "1/4",
  "t": 3
}
```
