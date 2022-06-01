SHELL := /bin/bash

SERVICE_NAME := quasar
IMAGE := gcr.io/${PROJECT_ID}/${SERVICE_NAME}
REGION := asia-northeast1

update:
	go get -u
	go mod tidy

test:
	GOOGLE_CLOUD_PROJECT=${PROJECT_ID} go test -v -coverprofile=coverage.out -covermode=atomic -coverpkg ./...

merge:
	echo "" > coverage.txt
	cat coverage.out     >> coverage.txt
	cat coverage-pkg.out >> coverage.txt

build:
	# gcloud builds submit --project ${PROJECT_ID} --tag ${IMAGE}

	gcloud auth configure-docker --quiet
	docker build -t ${IMAGE} .
	docker push ${IMAGE}

deploy:
	gcloud run deploy --region ${REGION} --project ${PROJECT_ID} --image ${IMAGE} --set-env-vars=GOOGLE_CLOUD_PROJECT=${PROJECT_ID},GIN_MODE=release ${SERVICE_NAME} 

package:
	docker tag ${IMAGE} ghcr.io/itsubaki/${SERVICE_NAME}
	docker push ghcr.io/itsubaki/${SERVICE_NAME}

up:
	echo "PROJECT_ID: ${PROJECT_ID}"
	echo "GOOGLE_APPLICATION_CREDENTIALS: ${GOOGLE_APPLICATION_CREDENTIALS}"
	docker-compose up

run:
	GOOGLE_CLOUD_PROJECT=${PROJECT_ID} USE_PPROF=true go run main.go

shor:
	curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" $(shell gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --format 'value(status.url)')/shor/15 | jq .

qasm:
	curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" $(shell gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --format 'value(status.url)') -X POST -F file=@testdata/shor.qasm | jq .
