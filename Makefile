SHELL := /bin/bash
DATE := $(shell date +%Y%m%d-%H:%M:%S)
HASH := $(shell git rev-parse HEAD)
GOVERSION := $(shell go version)
LDFLAGS := -X 'main.date=${DATE}' -X 'main.hash=${HASH}' -X 'main.goversion=${GOVERSION}'

install:
	-rm ${GOPATH}/bin/quasar
	go get -u
	go mod tidy
	go install -ldflags "${LDFLAGS}"

test:
	GOOGLE_APPLICATION_CREDENTIALS=../credentials.json go test --godog.format=pretty -v -coverprofile=coverage.out -covermode=atomic -coverpkg ./...

run:
	GOOGLE_APPLICATION_CREDENTIALS=./credentials.json go run main.go

merge:
	echo "" > coverage.txt
	cat coverage.out     >> coverage.txt
	cat coverage-pkg.out >> coverage.txt

deploy:
	gcloud app deploy app.yaml --project ${PROJECT_ID}

browse:
	gcloud app browse
