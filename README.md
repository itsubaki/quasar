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
$ curl -s -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://quasar-an.a.run.app/shor/15 | jq .
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
