SHELL := /bin/bash

install:
	-rm ${GOPATH}/bin/quasar
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
	echo "project: ${PROJECT_ID}"
	gcloud builds submit     --tag gcr.io/${PROJECT_ID}/quasar   --project ${PROJECT_ID}
	gcloud run deploy quasar --image gcr.io/${PROJECT_ID}/quasar --project ${PROJECT_ID} --set-env-vars=GOOGLE_CLOUD_PROJECT=${PROJECT_ID}

shor:
	curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" $(shell gcloud run services describe quasar --project ${PROJECT_ID} --format 'value(status.url)')/shor/15 | jq .

qasm:
	curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" $(shell gcloud run services describe quasar --project ${PROJECT_ID} --format 'value(status.url)') -X POST -F file=@testdata/shor.qasm | jq .
