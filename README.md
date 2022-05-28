# quasar

Quantum Computation Simulator as a Service

## Deploying to Cloud Run

```shell
$ export PROJECT_ID=YOUR_GOOGLE_CLOUD_PROJECT_ID
$ gcloud builds submit --tag gcr.io/${PROJECT_ID}/quasar
$ gcloud run deploy --image gcr.io/${PROJECT_ID}/quasar --set-env-vars=GOOGLE_CLOUD_PROJECT=${PROJECT_ID} quasar
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

$ curl -s $(gcloud run services describe quasar --format 'value(status.url)') -X POST -F file=@testdata/bell.qasm | jq .
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
```

```shell
$ curl -s -H $(gcloud run services describe quasar --format 'value(status.url)')/shor/15 | jq .
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
