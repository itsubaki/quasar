SHELL := /bin/bash

PROJECT_ID := $(shell gcloud config get-value project)
TARGET_URL := $(shell gcloud run services describe quasar --region asia-northeast1 --format 'value(status.url)' --project ${PROJECT_ID})
SERVICE_NAME := quasar
LOCATION := asia-northeast1
IMAGE := ${LOCATION}-docker.pkg.dev/${PROJECT_ID}/${SERVICE_NAME}/app
TAG := latest

update:
	GOPROXY=direct go get github.com/itsubaki/qasm@HEAD
	go get -u
	go mod tidy

install:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

.PHONY: gen
gen:
	rm -rf gen
	buf lint
	buf generate

test:
	PROJECT_ID=${PROJECT_ID} go test -v -coverprofile=coverage.txt -covermode=atomic -coverpkg=./...

testwip:
	PROJECT_ID=${PROJECT_ID} go test -v -godog.tags=wip -coverprofile=coverage.txt -covermode=atomic -coverpkg=./...

testpkg:
	PROJECT_ID=${PROJECT_ID} go test -v -cover $(shell go list ./... | grep -v /vendor/ | grep -v /build/ | grep -v -E "quasar$$") -coverprofile=coverage-pkg.txt -covermode=atomic

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
	gcloud run deploy --region ${LOCATION} --project ${PROJECT_ID} --image ${IMAGE} --set-env-vars=PROJECT_ID=${PROJECT_ID},USE_CPROF=true,MAX_QUBITS=10 ${SERVICE_NAME}
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

bell:
	@curl -s \
		-H "Authorization: Bearer $(shell gcloud auth print-identity-token)" \
		-H "Content-Type: application/json" \
		-d '{"code": "OPENQASM 3.0; gate h q { U(pi/2.0, 0, pi) q; } gate x q { U(pi, 0, pi) q; } gate cx c, t { ctrl @ U(pi, 0, pi) c, t; } qubit[2] q; reset q; h q[0]; cx q[0], q[1];"}' \
		$(shell gcloud run services describe ${SERVICE_NAME} --project ${PROJECT_ID} --format 'value(status.url)')/quasar.v1.QuasarService/Simulate | jq .

curl:
	@curl -s \
		-H 'Content-Type: application/json' \
		-d '{"code": "OPENQASM 3.0; gate h q { U(pi/2.0, 0, pi) q; } gate x q { U(pi, 0, pi) q; } gate cx c, t { ctrl @ U(pi, 0, pi) c, t; } qubit[2] q; reset q; h q[0]; cx q[0], q[1];"}' \
		localhost:8080/quasar.v1.QuasarService/Simulate | jq .

save:
	@curl -s \
		-H 'Content-Type: application/json' \
		-d '{"code": "OPENQASM 3.0; gate h q { U(pi/2.0, 0, pi) q; } gate x q { U(pi, 0, pi) q; } gate cx c, t { ctrl @ U(pi, 0, pi) c, t; } qubit[2] q; reset q; h q[0]; cx q[0], q[1];"}' \
		localhost:8080/quasar.v1.QuasarService/Save | jq .

load:
	@curl -s \
		-H 'Content-Type: application/json' \
		-d '{"id": "V1hz_VNsg68hcldd"}' \
		localhost:8080/quasar.v1.QuasarService/Load | jq .
