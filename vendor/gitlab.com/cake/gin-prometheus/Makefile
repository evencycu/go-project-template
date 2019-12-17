APP=gin-metrics
PWD=$(shell pwd)
SOURCE=./example/...
GOPATH=$(shell env | grep GOPATH | cut -d'=' -f 2)

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
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

