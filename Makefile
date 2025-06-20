SHELL := /bin/bash

PROJECT_ID := $(shell gcloud config get-value project)
TARGET_URL := $(shell gcloud run services describe quasar --region asia-northeast1 --format 'value(status.url)' --project ${PROJECT_ID})
SERVICE_NAME := quasar
LOCATION := asia-northeast1
IMAGE := ${LOCATION}-docker.pkg.dev/${PROJECT_ID}/${SERVICE_NAME}/app
TAG := latest

update:
	go get -u
	go mod tidy

test:
	PROJECT_ID=${PROJECT_ID} LOG_LEVEL=5 go test -v -coverprofile=coverage.txt -covermode=atomic -coverpkg ./...

testwip:
	PROJECT_ID=${PROJECT_ID} LOG_LEVEL=5 go test -v -godog.tags=wip -coverprofile=coverage.txt -covermode=atomic -coverpkg ./...

testpkg:
	PROJECT_ID=${PROJECT_ID} LOG_LEVEL=5 go test -v -cover $(shell go list ./... | grep -v /vendor/ | grep -v /build/ | grep -v -E "quasar$$") -coverprofile=coverage-pkg.txt -covermode=atomic

artifact:
	gcloud artifacts repositories create ${SERVICE_NAME} --repository-format=docker --location=${LOCATION} --project=${PROJECT_ID}

cloudbuild:
	gcloud builds submit --config cloudbuild.yaml --substitutions=_IMAGE=${IMAGE},_TAG=${TAG} .

build:
	gcloud auth configure-docker ${LOCATION}-docker.pkg.dev --quiet
	gcloud artifacts repositories list
	docker build -t ${IMAGE} .
	docker push ${IMAGE}

deploy:
	gcloud artifacts docker images describe ${IMAGE}
	gcloud run deploy --region ${LOCATION} --project ${PROJECT_ID} --image ${IMAGE} --set-env-vars=PROJECT_ID=${PROJECT_ID},USE_CPROF=true,GIN_MODE=release ${SERVICE_NAME}
	gcloud run services update-traffic ${SERVICE_NAME} --to-latest --region ${LOCATION} --project ${PROJECT_ID}

package:
	@echo ${PAT} | docker login ghcr.io -u itsubaki --password-stdin
	docker tag ${IMAGE} ghcr.io/itsubaki/${SERVICE_NAME}
	docker push ghcr.io/itsubaki/${SERVICE_NAME}

up:
	echo "PROJECT_ID: ${PROJECT_ID}"
	echo "GOOGLE_APPLICATION_CREDENTIALS: ${GOOGLE_APPLICATION_CREDENTIALS}"
	docker compose up

run:
	PROJECT_ID=${PROJECT_ID} USE_PPROF=true go run main.go

factorize:
	@curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" $(shell gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --format 'value(status.url)')/factorize/15 | jq .

qasm:
	@curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" $(shell gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --format 'value(status.url)') -X POST -F file=@testdata/qft.qasm | jq .
