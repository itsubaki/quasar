# quasar

Quantum Computation Simulator as a Service

## Deployment

```shell
$ gcloud app deploy app.yaml
```

## Example

```shell
$ curl -s https://${PROJECT_ID}.an.r.appspot.com/shor/15 | jq .
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
