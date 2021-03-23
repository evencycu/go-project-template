APP=ginmetrics
PWD=$(shell pwd)
SOURCE=./...
GOPATH=$(shell env | grep GOPATH | cut -d'=' -f 2)
export GOPRIVATE=gitlab.com*

build:
	GO111MODULE=on go install -v $(SOURCE)

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover $(SOURCE)

run-origin:build
	$(GOPATH)/bin/gin-metrics-origin

run-custom:build
	$(GOPATH)/bin/gin-metrics-custom

modvendor:
	GO111MODULE=on go build -mod=mod -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

