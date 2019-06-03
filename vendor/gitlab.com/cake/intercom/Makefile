PWD=$(shell pwd)
SOURCE=./...
GOPATH=$(shell env | grep GOPATH | cut -d'=' -f 2)
export GO111MODULE=on

build: 
	go build -v $(SOURCE) 

test:
	@echo "Start unit tests & vet..."
	go vet $(SOURCE)
	go test -race -cover $(SOURCE)

clean:
	rm -rf bin pkg
	docker system prune -f
	go mod tidy

modvendor:
	GO111MODULE=on go build -v $(SOURCE) 
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor