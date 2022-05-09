SHELL := /bin/bash

SERVICE_NAME := quasar
IMAGE := gcr.io/${PROJECT_ID}/${SERVICE_NAME}
TOKEN := $(shell gcloud auth print-identity-token)
URL   := $(shell gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --format 'value(status.url)')

update:
	go get -u
	go mod tidy

test:
	go test --godog.format=pretty -v -coverprofile=coverage.out -covermode=atomic -coverpkg ./...

run:
	go run main.go

merge:
	echo "" > coverage.txt
	cat coverage.out     >> coverage.txt
	cat coverage-pkg.out >> coverage.txt

deploy:
	echo "PROJECT_ID=${PROJECT_ID}"
	gcloud builds submit --project ${PROJECT_ID} --tag   ${IMAGE} 
	gcloud run deploy --project ${PROJECT_ID} --image ${IMAGE} --set-env-vars=GOOGLE_CLOUD_PROJECT=${PROJECT_ID},GIN_MODE=release ${SERVICE_NAME} 

shor:
	curl -s -H "Authorization: Bearer ${TOKEN}" ${URL}/shor/15 | jq .

qasm:
	curl -s -H "Authorization: Bearer ${TOKEN}" ${URL} -X POST -F file=@testdata/shor.qasm | jq .
