[![PkgGoDev](https://pkg.go.dev/badge/github.com/itsubaki/quasar)](https://pkg.go.dev/github.com/itsubaki/quasar)
[![Go Report Card](https://goreportcard.com/badge/github.com/itsubaki/quasar?style=flat-square)](https://goreportcard.com/report/github.com/itsubaki/quasar)
[![deploy](https://github.com/itsubaki/quasar/workflows/deploy/badge.svg)](https://github.com/itsubaki/quasar/actions)

# quasar

Quantum Computation Simulator as a Service

## Deploying to Cloud Run

```shell
$ export PROJECT_ID=YOUR_GOOGLE_CLOUD_PROJECT_ID
$ export LOCATION=YOUR_GOOGLE_CLOUD_LOCATION
$ export IMAGE=${LOCATION}-docker.pkg.dev/${PROJECT_ID}/quasar/app
$
$ gcloud builds submit --config cloudbuild.yaml --substitutions=_IMAGE=${IMAGE},_TAG=latest .
$ gcloud run deploy --image ${IMAGE} --set-env-vars=PROJECT_ID=${PROJECT_ID} quasar
```

## Examples

```shell
$ cat testdata/bell.qasm
OPENQASM 3.0;

gate h q { U(pi/2.0, 0, pi) q; }
gate x q { U(pi, 0, pi) q; }
gate cx c, t { ctrl @ U(pi, 0, pi) c, t; }

qubit[2] q;
reset q;

h q[0];
cx q[0], q[1];
```

```shell
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
  "seed": -1,
  "t": 3
}
```
