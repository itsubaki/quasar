SHELL := /bin/bash

install:
	-rm ${GOPATH}/bin/quasar
	go get -u
	go mod tidy

test:
	GOOGLE_APPLICATION_CREDENTIALS=../credentials.json go test --godog.format=pretty -v -coverprofile=coverage.out -covermode=atomic -coverpkg ./...

run:
	GOOGLE_APPLICATION_CREDENTIALS=./credentials.json go run main.go

merge:
	echo "" > coverage.txt
	cat coverage.out     >> coverage.txt
	cat coverage-pkg.out >> coverage.txt

deploy:
	echo "project: ${PROJECT_ID}"
	gcloud builds submit     --tag gcr.io/${PROJECT_ID}/quasar   --project ${PROJECT_ID}
	gcloud run deploy quasar --image gcr.io/${PROJECT_ID}/quasar --project ${PROJECT_ID} --set-env-vars=GOOGLE_CLOUD_PROJECT=${PROJECT_ID}

shor:
	echo "DOMAIN: ${CLOUDRUN_DOMAIN}"
	curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" https://${CLOUDRUN_DOMAIN}/shor/15 | jq .

qasm:
	echo "DOMAIN: ${CLOUDRUN_DOMAIN}"
	curl -s -H "Authorization: Bearer $(shell gcloud auth print-identity-token)" -X POST -F file=@testdata/shor.qasm https://${CLOUDRUN_DOMAIN} | jq .
