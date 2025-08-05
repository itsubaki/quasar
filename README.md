[![PkgGoDev](https://pkg.go.dev/badge/github.com/itsubaki/quasar)](https://pkg.go.dev/github.com/itsubaki/quasar)
[![Go Report Card](https://goreportcard.com/badge/github.com/itsubaki/quasar?style=flat-square)](https://goreportcard.com/report/github.com/itsubaki/quasar)
[![deploy](https://github.com/itsubaki/quasar/workflows/deploy/badge.svg)](https://github.com/itsubaki/quasar/actions)

# quasar

- Quantum computation simulator as a Service

## Deploying to Cloud Run

```shell
export PROJECT_ID=YOUR_GOOGLE_CLOUD_PROJECT_ID
export LOCATION=YOUR_GOOGLE_CLOUD_LOCATION
export IMAGE=${LOCATION}-docker.pkg.dev/${PROJECT_ID}/quasar/app

gcloud builds submit --config cloudbuild.yaml --substitutions=_IMAGE=${IMAGE},_TAG=latest .
gcloud run deploy --image ${IMAGE} --set-env-vars=PROJECT_ID=${PROJECT_ID} quasar
```

## Examples

```shell
curl -s \
    $(gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --format 'value(status.url)')/quasar.v1.QuasarService/Simulate 
    -H "Authorization: Bearer $(gcloud auth print-identity-token)" \
    -H "Content-Type: application/json" \
    -d '{"code": "OPENQASM 3.0; gate h q { U(pi/2.0, 0, pi) q; } gate x q { U(pi, 0, pi) q; } gate cx c, t { ctrl @ U(pi, 0, pi) c, t; } qubit[2] q; reset q; h q[0]; cx q[0], q[1];"}' | jq
```

```json
{
  "state": [
    {
      "amplitude": {
        "real": 0.7071,
        "imag": 0
      },
      "probability": 0.5,
      "int": [
        0
      ],
      "binary_string": [
        "00"
      ]
    },
    {
      "amplitude": {
        "real": 0.7071,
        "imag": 0
      },
      "probability": 0.5,
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
